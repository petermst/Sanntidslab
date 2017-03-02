package main

import (
	. "./Driver"
	. "./FSM"
	"fmt"
	"os"
	//"time"
)

func main() {
	InitializeElevator() //Skal stå

	fmt.Println("Press STOP button to stop elevator and exit program.\n")

	ElevSetMotorDirection(0)

	setButtonIndicator := make(chan ButtonIndicator, 1) //Driver <- Queue
	setMotorDirection := make(chan int, 1)              //Driver <- FSM
	startDoorTimer := make(chan bool, 1)                //Driver <- FSM
	eventElevatorStuck := make(chan bool, 1)            //FSM <- Driver
	eventAtFloor := make(chan int, 1)                   //FSM <- Driver
	eventDoorTimeout := make(chan bool, 1)              //FSM <- Driver
	nextDirection := make(chan int, 1)                  //FSM <- Queue
	shouldStop := make(chan int, 1)                     //Queue <- FSM
	getNextDirection := make(chan bool, 1)              //Queue <- FSM
	elevatorStuckUpdateQueue := make(chan bool, 1)      //Queue <- FSM
	//doorTimer := make(chan time.Time)

	go RunDriver(setButtonIndicator, setMotorDirection, startDoorTimer, elevatorStuck, eventAtFloor, eventDoorTimeout)
	go RunFSM(setMotorDirection, startDoorTimer, elevatorStuck, eventAtFloor, nextDir, shouldStop, getNextDirection, elevatorStuckUpdateQueue)

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
