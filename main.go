package main

import (
	"fmt"
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

	tick := time.Tick(1 * time.Second)

	for range tick {
		str := fmt.Sprintf("{ \"time\": %d,", time.Now().Unix())

		str += `"load": `
		load, err := obd.GetEngineLoad()
		if err != nil {
			str += "null"
		} else {
			str += fmt.Sprintf("%f", load)
		}

		str += `, "temp": `
		temp, err := obd.GetCoolantTemp()
		if err != nil {
			str += "null"
		} else {
			str += fmt.Sprintf("%d", temp)
		}

		str += `, "rpm": `
		rpm, err := obd.GetRPM()
		if err != nil {
			str += "null"
		} else {
			str += fmt.Sprintf("%f", rpm)
		}

		str += `, "speed": `
		speed, err := obd.GetSpeed()
		if err != nil {
			str += "null"
		} else {
			str += fmt.Sprintf("%d", speed)
		}

		str += `, "throttle": `
		throt, err := obd.GetThrottlePosition()
		if err != nil {
			str += "null"
		} else {
			str += fmt.Sprintf("%f", throt)
		}

		str += "}"

		log.Println(str)

	}

}
