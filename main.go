package main

import (
	. "./Driver"
	"fmt"
	"os"
	"time"
)

func main() {
	ElevInit() //Skal stå

	fmt.Println("Press STOP button to stop elevator and exit program.\n")

	ElevSetMotorDirection(0)

	setButtonLamp := make(chan buttonIndicator)
	setMotorDirection := make(chan int)
	startDoorTimer := make(chan bool)
	//doorTimer := make(chan time.Time)

	RunDriver(setButtonLamp, setMotorDirection, startDoorTimer)

	for {
		// Change direction when we reach top/bottom floor
		/*if ElevGetFloorSensorSignal() == (4 - 1) {
			ElevSetMotorDirection(-1)
		} else if ElevGetFloorSensorSignal() == 0 {
			ElevSetMotorDirection(1)
		}*/

		// Stop elevator and exit program if the stop button is pressed
		if ElevGetStopSignal() != 0 {
			ElevSetMotorDirection(0)
			os.Exit(1)
		}
	}
}
