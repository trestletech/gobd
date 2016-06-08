package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

type OBD struct {
	pids  []uint
	ser   *serial.Port
	debug func(format string, v ...interface{})
	id    string
}

func NewOBD(ser *serial.Port) (*OBD, error) {
	return NewDebugOBD(ser, func(string, ...interface{}) {})
}

func NewDebugOBD(ser *serial.Port, debug func(format string, v ...interface{})) (*OBD, error) {
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
	_, err = obd.exec("ATE0")
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

var engineLoadPID int = 4

func (o *OBD) GetEngineLoad() (float64, error) {
	if !includes(o.pids, engineLoadPID) {
		return float64(0), pidNotSupportedError(engineLoadPID)
	}
	val, err := o.currentInt(engineLoadPID)
	return float64(val) / 2.55, err
}

func (o *OBD) current(pid int) ([]byte, error) {
	cmd := fmt.Sprintf("01%02x\r\n", pid)
	log.Printf("Writing %s", cmd)
	n, err := o.ser.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 128)
	n, err = o.ser.Read(buf)
	log.Printf("Read: %s", string(buf[:n]))
	refined := refineInput(buf[:n])
	//log.Printf("Refined: %v", refined)
	return refined, err
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
	allPids, err := obd.listPIDsForBase(0)
	if err != nil {
		return allPids, err
	}
	if !includes(allPids, 32) {
		return allPids, nil
	}

	pids, err := obd.listPIDsForBase(32)
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
	cmd := fmt.Sprintf("01%02x\r\n", base)
	//log.Printf("Polling for base %d", base)
	_, err := obd.ser.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 128)
	n, err := obd.ser.Read(buf)
	if err != nil {
		return nil, err
	}

	//log.Printf("Read: %s", string(buf[:n]))

	return extractPids(buf[:n]), nil
}

func extractPids(buf []byte) []uint {
	buf = refineInput(buf)

	//log.Printf("Working with: %s", string(buf))

	pids := make([]uint, 0)
	for bytCount := 0; bytCount < len(buf); bytCount++ {
		byt, err := strconv.ParseInt(string(buf[bytCount]), 16, 8)
		if err != nil {
			log.Fatalf("Can't parse hex %s in %v", string(buf[bytCount]), buf)
		}
		for i := 0; i < 4; i++ {
			if byt >= 8 {
				pids = append(pids, uint(bytCount*4+i+1))
			}
			byt <<= 1
			byt = byt % 16
		}
	}
	return pids
}

func refineInput(buf []byte) []byte {
	// trim spaces
	spaces := []byte{byte(' ')}
	spl := bytes.Split(buf, spaces)
	buf = bytes.Join(spl, nil)

	newline := bytes.IndexAny(buf, "\n\r")
	if newline < 0 {
		log.Fatal("Error reading.")
	}
	// TODO: confirm instead of throwing away
	//log.Printf("Buf: %+v", buf)
	//log.Printf("newline: %d", newline)
	return buf[4:newline]
}
