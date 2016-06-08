package main

import (
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	c := &serial.Config{Name: "/dev/tty.usbmodem1421", Baud: 9600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	obd, err := NewDebugOBD(s, log.Printf)
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
