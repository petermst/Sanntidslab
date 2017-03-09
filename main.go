package main

import (
	. "./Driver"
	. "./FSM"
	. "./Network"
	. "./Queue"
	"fmt"
	"os"
	//"time"
)

func main() {
	InitializeElevator() //Skal stå

	fmt.Println("Press STOP button to stop elevator and exit program.\n")

	ElevSetMotorDirection(0)

	const id = InitializeNetwork()

	setMotorDirection := make(chan int, 1)              //Driver <- FSM
	startDoorTimer := make(chan bool, 1)                //Driver <- FSM
	setButtonIndicator := make(chan ButtonIndicator, 1) //Driver <- Queue
	eventElevatorStuck := make(chan bool, 1)            //FSM <- Driver
	eventAtFloor := make(chan int, 1)                   //FSM <- Driver
	eventDoorTimeout := make(chan bool, 1)              //FSM <- Driver
	nextDirection := make(chan int, 1)                  //FSM <- Queue
	calcOptimalElevator := make(chan Order, 1)          //Queue <- Driver
	shouldStop := make(chan int, 1)                     //Queue <- FSM
	getNextDirection := make(chan bool, 1)              //Queue <- FSM
	elevatorStuckUpdateQueue := make(chan bool, 1)      //Queue <- FSM
	updateQueue := make(chan QueueOperation, 1)         //Queue <- FSM
	updatePeersOnQueue := make(chan driverState, 1)     //Queue <- Network
	messageSent := make(chan QueueOperation)            //Queue <- Network
	updateQueueSize := make(chan NewOrLostPeer, 1)      //Queue <- Network
	incomingMSG := make(chan QueueOperation)            //Queue <- Network
	outgoingMSG := make(chan QueueOperation)            //Network <- Queue
	peersTransmitMSG := make(chan driverState)          //Network <- Queue

	go RunDriver(setButtonIndicator, setMotorDirection, startDoorTimer, eventElevatorStuck, eventAtFloor, eventDoorTimeout)
	go RunFSM(id, setMotorDirection, startDoorTimer, eventElevatorStuck, eventAtFloor, nextDirection, shouldStop, getNextDirection, elevatorStuckUpdateQueue, updateQueue)
	go RunNetwork(id, updatePeersOnQueue, updateQueueSize, incomingMSG, outgoingMSG, peersTransmitMSG, messageSent)
	go RunQueue(id, updateQueue, updateQueueSize, shouldStop, setButtonIndicator, incomingMSG, outgoingMSG, messageSent, nextDirection, getNextDirection, peersTransmitMSG, elevatorStuckUpdateQueue)

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

/*
Må fikse å sette lys fra kø-modulen (spesielt med tanke på interne vs. eksterne ordre).
Må fikse sletting av ordre fra UpdateQueue() mtp. interne ordre vs. eksterne. (en annen heis ankommer skal ikke slette interne)
Fikse at Network sier ifra til Queue når den har sendt en melding, slik at den kan sette lys.
Fikse incoming og outgoing messages(grensesnitt mot Network fra Queue)








*/
