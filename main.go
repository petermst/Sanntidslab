package main

import (
	. "./Def"
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

	setMotorDirectionCh := make(chan int, 1)              //Driver <- FSM
	startDoorTimerCh := make(chan bool, 1)                //Driver <- FSM
	setButtonIndicatorCh := make(chan ButtonIndicator, 1) //Driver <- Queue
	eventElevatorStuckCh := make(chan bool, 1)            //FSM <- Driver
	eventAtFloorCh := make(chan int, 1)                   //FSM <- Driver
	eventDoorTimeoutCh := make(chan bool, 1)              //FSM <- Driver
	nextDirectionCh := make(chan int, 1)                  //FSM <- Queue
	calcOptimalElevatorCh := make(chan Order, 1)          //Queue <- Driver
	shouldStopCh := make(chan int, 1)                     //Queue <- FSM
	getNextDirectionCh := make(chan bool, 1)              //Queue <- FSM
	elevatorStuckUpdateQueueCh := make(chan bool, 1)      //Queue <- FSM
	updateQueueCh := make(chan QueueOperation, 1)         //Queue <- FSM
	updatePeersOnQueueCh := make(chan driverState, 1)     //Queue <- Network
	messageSentCh := make(chan QueueOperation, 1)         //Queue <- Network
	updateQueueSizeCh := make(chan NewOrLostPeer, 1)      //Queue <- Network
	incomingMSGCh := make(chan QueueOperation, 1)         //Queue <- Network
	outgoingMSGCh := make(chan QueueOperation, 1)         //Network <- Queue
	peersTransmitMSGCh := make(chan driverState, 1)       //Network <- Queue

	go RunDriver(id, setButtonIndicatorCh, setMotorDirectionCh, startDoorTimerCh, eventElevatorStuckCh, eventAtFloorCh, eventDoorTimeoutCh)
	go RunFSM(id, setMotorDirectionCh, startDoorTimerCh, eventElevatorStuckCh, eventAtFloorCh, nextDirectionCh, shouldStopCh, getNextDirectionCh, elevatorStuckUpdateQueueCh, updateQueueCh)
	go RunNetwork(id, updatePeersOnQueueCh, updateQueueSizeCh, incomingMSGCh, outgoingMSGCh, peersTransmitMSGCh, messageSentCh)
	go RunQueue(id, calcOptimalElevatorCh, updateQueueCh, updateQueueSizeCh, shouldStopCh, setButtonIndicatorCh, incomingMSGCh, outgoingMSGCh, messageSentCh, nextDirectionCh, getNextDirectionCh, peersTransmitMSGCh, elevatorStuckUpdateQueueCh)

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
