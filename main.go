package main

import (
	. "./Driver"
	"fmt"
	"os"
)

func main() {
	ElevInit()

	fmt.Println("Press STOP button to stop elevator and exit program.\n")

	ElevSetMotorDirection(1)

	for {
		// Change direction when we reach top/bottom floor
		if ElevGetFloorSensorSignal() == (4 - 1) {
			ElevSetMotorDirection(-1)
		} else if ElevGetFloorSensorSignal() == 0 {
			ElevSetMotorDirection(1)
		}

		// Stop elevator and exit program if the stop button is pressed
		if ElevGetStopSignal() != 0 {
			ElevSetMotorDirection(0)
			os.Exit(1)
		}
	}
}
