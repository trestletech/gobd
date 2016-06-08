package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
)

type OBD struct {
	pids  []uint
	ser   SerialPort
	debug func(format string, v ...interface{})
	id    string
}

func NewOBD(ser SerialPort) (*OBD, error) {
	return NewDebugOBD(ser, func(string, ...interface{}) {})
}

func NewDebugOBD(ser SerialPort, debug func(format string, v ...interface{})) (*OBD, error) {
	obd := &OBD{
		ser:   ser,
		debug: debug,
	}

	// Reset
	// We don't use odb.exec here because we need to insert the time.Sleep
	obd.debug("Resetting...")
	err := obd.execNoRead("ATZ")
	if err != nil {
		return obd, err
	}

	// Important to avoid reading something already in the buffer and not
	// properly blocking until the output of the previous command comes through.
	time.Sleep(500 * time.Millisecond)

	id, err := obd.read()
	if err != nil {
		return obd, err
	}
	obd.id = string(id)

	// Disable Echo
	obd.debug("Disabling echo...")
	err = obd.execNoRead("ATE0")
	if err != nil {
		return obd, err
	}

	time.Sleep(100 * time.Millisecond)

	_, err = obd.read()
	if err != nil {
		return obd, err
	}

	obd.debug("Getting available PIDs...")
	pids, err := obd.listSupportedPIDs()
	if err != nil {
		return &OBD{}, err
	}
	obd.debug("PIDs: %v", pids)
	obd.pids = pids
	obd.debug("\tDone.")
	return obd, nil
}

// Execute an arbitrary command by sending it over the serial port.
func (obd *OBD) exec(cmd string) ([]byte, error) {
	err := obd.execNoRead(cmd)
	if err != nil {
		return nil, err
	}

	return obd.read()
}
func (obd *OBD) execNoRead(cmd string) error {
	obd.debug("Executing '%s'...", cmd)
	_, err := obd.ser.Write([]byte(cmd + "\r\n"))
	return err
}

func (obd *OBD) read() ([]byte, error) {
	buf := make([]byte, 128)
	n, err := obd.ser.Read(buf)
	if err != nil {
		return nil, err
	}
	out := parseOutput(buf[:n])
	obd.debug("\t%s", string(out))
	obd.debug("\tDone.")
	return out, nil
}

func parseOutput(out []byte) []byte {
	toRet := bytes.TrimLeft(out, "\n\r >")
	toRet = bytes.TrimRight(toRet, "\n\r >")
	return toRet
}

func pidNotSupportedError(pid int) error {
	return fmt.Errorf("PID %d not supported on this system.", pid)
}

func (obd *OBD) current(pid int) ([]byte, error) {
	cmd := fmt.Sprintf("01%02x", pid)
	out, err := obd.exec(cmd)
	if err != nil {
		return nil, err
	}

	return parseMode1Response(out)
}

func (o *OBD) currentInt(pid int) (int, error) {
	b, err := o.current(pid)
	if err != nil {
		return 0, err
	}

	val := 0
	factor := 1
	for i := len(b) - 1; i >= 0; i-- {
		byt, err := strconv.ParseInt(string(b[i]), 16, 8)
		if err != nil {
			log.Fatalf("Can't parse hex %s in %v", string(b[i]), b)
		}

		val += int(byt) * factor
		factor *= 16
	}

	return val, nil
}

func includes(haystack []uint, needle int) bool {
	for _, p := range haystack {
		if uint(needle) == p {
			return true
		}
	}
	return false
}

func (obd *OBD) listSupportedPIDs() ([]uint, error) {
	allPids := []uint{0}
	pids, err := obd.listPIDsForBase(0)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 32) {
		return allPids, nil
	}

	pids, err = obd.listPIDsForBase(32)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 64) {
		return allPids, nil
	}

	pids, err = obd.listPIDsForBase(64)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 96) {
		return allPids, nil
	}

	pids, err = obd.listPIDsForBase(96)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 128) {
		return allPids, nil
	}

	pids, err = obd.listPIDsForBase(128)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 160) {
		return allPids, nil
	}

	pids, err = obd.listPIDsForBase(160)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 192) {
		return allPids, nil
	}

	pids, err = obd.listPIDsForBase(192)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)

	return allPids, nil
}

func (obd *OBD) listPIDsForBase(base int) ([]uint, error) {
	obd.debug("Getting PIDs %d", base)
	cmd := fmt.Sprintf("01%02x", base)
	out, err := obd.exec(cmd)
	if err != nil {
		return nil, err
	}

	res, err := parseMode1Response(out)
	if err != nil {
		return nil, err
	}
	return extractPids(res, base), nil
}

func extractPids(buf []byte, base int) []uint {
	pids := make([]uint, 0)
	for bytCount := 0; bytCount < len(buf); bytCount++ {
		byt, err := strconv.ParseInt(string(buf[bytCount]), 16, 8)
		if err != nil {
			log.Fatalf("Can't parse hex %s in %v", string(buf[bytCount]), buf)
		}
		for i := 0; i < 4; i++ {
			if byt >= 8 {
				pids = append(pids, uint(base+bytCount*4+i+1))
			}
			byt <<= 1
			byt = byt % 16
		}
	}
	return pids
}

func parseMode1Response(buf []byte) ([]byte, error) {
	if len(buf) < 2 || buf[0] != '4' || buf[1] != '1' {
		// Error response
		return nil, fmt.Errorf("Error mode 1 prefix response: %s", string(buf))
	}

	// trim spaces
	spaces := []byte{byte(' ')}
	spl := bytes.Split(buf, spaces)
	buf = bytes.Join(spl, nil)

	return buf[4:], nil
}
