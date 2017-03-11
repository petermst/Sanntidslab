package main

import (
	. "./Def"
	. "./Driver"
	. "./FSM"
	. "./Network"
	. "./Queue"
	"fmt"
)

func main() {
	initFloor := InitializeElevator()
	fmt.Printf("initfloor er: %d\n", initFloor)

	id := InitializeNetwork()

	fmt.Printf("Dette er min ID: %s\n", id)

	setMotorDirectionCh := make(chan int, 1)                 //Driver <- FSM
	startDoorTimerCh := make(chan bool, 1)                   //Driver <- FSM
	setButtonIndicatorCh := make(chan ButtonIndicator, 1)    //Driver <- Queue
	eventElevatorStuckCh := make(chan bool, 1)               //FSM <- Driver
	eventAtFloorCh := make(chan int, 1)                      //FSM <- Driver
	eventDoorTimeoutCh := make(chan bool, 1)                 //FSM <- Driver
	nextDirectionCh := make(chan []int, 1)                   //FSM <- Queue
	calcOptimalElevatorCh := make(chan Order, 1)             //Queue <- Driver
	shouldStopCh := make(chan int, 1)                        //Queue <- FSM
	getNextDirectionCh := make(chan bool, 1)                 //Queue <- FSM
	elevatorStuckUpdateQueueCh := make(chan bool, 1)         //Queue <- FSM
	updateQueueCh := make(chan QueueOperation, 1)            //Queue <- Driver
	messageSentCh := make(chan QueueOperation, 1)            //Queue <- Network
	updateQueueSizeCh := make(chan NewOrLostPeer, 1)         //Queue <- Network
	incomingQueueUpdateCh := make(chan QueueOperation, 1)    //Queue <- Network
	outgoingQueueUpdateCh := make(chan QueueOperation, 1)    //Network <- Queue
	incomingDriverStateUpdateCh := make(chan DriverState, 1) //Queue <- Network
	outgoingDriverStateUpdateCh := make(chan DriverState, 1) //Network <- Queue
	isDoorOpenCh := make(chan bool, 1)                       //Driver <- Queue
	isDoorOpenResponseCh := make(chan bool, 1)               //Queue <- Driver

	go RunDriver(id, setButtonIndicatorCh, setMotorDirectionCh, startDoorTimerCh, eventElevatorStuckCh, eventAtFloorCh, eventDoorTimeoutCh, calcOptimalElevatorCh, updateQueueCh, isDoorOpenCh, isDoorOpenResponseCh)
	go RunFSM(id, setMotorDirectionCh, startDoorTimerCh, eventElevatorStuckCh, eventAtFloorCh, nextDirectionCh, shouldStopCh, eventDoorTimeoutCh, getNextDirectionCh, elevatorStuckUpdateQueueCh)
	go RunNetwork(id, updateQueueSizeCh, messageSentCh, outgoingQueueUpdateCh, incomingQueueUpdateCh, outgoingDriverStateUpdateCh, incomingDriverStateUpdateCh)
	go RunQueue(id, initFloor, calcOptimalElevatorCh, updateQueueCh, updateQueueSizeCh, shouldStopCh, setButtonIndicatorCh, messageSentCh, nextDirectionCh, getNextDirectionCh, elevatorStuckUpdateQueueCh, outgoingQueueUpdateCh, incomingQueueUpdateCh, outgoingDriverStateUpdateCh, incomingDriverStateUpdateCh, isDoorOpenCh, isDoorOpenResponseCh)

	/*
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
	*/
	stopElevator := make(chan bool)
	<-stopElevator
}

/*
Må fikse å sette lys fra kø-modulen (spesielt med tanke på interne vs. eksterne ordre).
Må fikse sletting av ordre fra UpdateQueue() mtp. interne ordre vs. eksterne. (en annen heis ankommer skal ikke slette interne)
Fikse at Network sier ifra til Queue når den har sendt en melding, slik at den kan sette lys.
Fikse incoming og outgoing messages(grensesnitt mot Network fra Queue)








*/
