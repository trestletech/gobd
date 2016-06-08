package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

func main() {
	c := &serial.Config{Name: "/dev/tty.usbmodem1411", Baud: 9600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	obd, err := NewOBD(s)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Supported PIDs: %v", obd.pids)

	load, err := obd.GetEngineLoad()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Engine load: %f", load)

}

type OBD struct {
	pids []uint
	ser  *serial.Port
}

func NewOBD(ser *serial.Port) (*OBD, error) {
	pids, err := listSupportedPIDs(ser)
	if err != nil {
		return &OBD{}, err
	}
	obd := OBD{
		pids: pids,
		ser:  ser,
	}

	// Reset
	_, err = ser.Write([]byte("ATWS\r\n"))
	if err != nil {
		return &obd, err
	}
	buf := make([]byte, 128)
	n, err := ser.Read(buf)
	if err != nil {
		return &obd, err
	}
	log.Printf("Reset output: %s", string(buf[:n]))

	// Disable Echo
	_, err = ser.Write([]byte("ATE0\r\n"))
	if err != nil {
		return &obd, err
	}

	time.Sleep(100 * time.Millisecond)
	buf = make([]byte, 128)
	n, err = ser.Read(buf)
	if err != nil {
		return &obd, err
	}
	log.Printf("Disable echo output: %s", string(buf[:n]))

	return &obd, nil
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

func listSupportedPIDs(ser *serial.Port) ([]uint, error) {
	allPids, err := listPIDsForBase(ser, 0)
	if err != nil {
		return allPids, err
	}
	if !includes(allPids, 32) {
		return allPids, nil
	}

	pids, err := listPIDsForBase(ser, 32)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 64) {
		return allPids, nil
	}

	pids, err = listPIDsForBase(ser, 64)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 96) {
		return allPids, nil
	}

	pids, err = listPIDsForBase(ser, 96)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 128) {
		return allPids, nil
	}

	pids, err = listPIDsForBase(ser, 128)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 160) {
		return allPids, nil
	}

	pids, err = listPIDsForBase(ser, 160)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)
	if !includes(allPids, 192) {
		return allPids, nil
	}

	pids, err = listPIDsForBase(ser, 192)
	if err != nil {
		return allPids, err
	}
	allPids = append(allPids, pids...)

	return allPids, nil
}

func listPIDsForBase(ser *serial.Port, base int) ([]uint, error) {
	cmd := fmt.Sprintf("01%02x\r\n", base)
	//log.Printf("Polling for base %d", base)
	_, err := ser.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 128)
	n, err := ser.Read(buf)
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
