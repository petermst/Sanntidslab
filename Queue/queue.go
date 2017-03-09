package Queue

import (
	. "../Driver"
	. "../FSM"
	. "../Network"
	"fmt"
	//"os"
	"time"
)

var n_elevators = 1

type QueueOperation struct {
	isAddOrder 	bool
	elevatorId  string
	floor     	int
	button 		int
}

type Order struct {
	floor  int
	button int
}

type QueueMap struct {
	mux   sync.Mutex
	queue map[string][][]bool
}

type DriverStatesMap struct {
	mux sync.Mutex
	states map[string][]int
}

func RunQueue(id string,calcOptimalElevator <-chan Order, updateQueue <-chan QueueOperation, updateQueueSize <-chan NewOrLostPeer, shouldStop <-chan int, setButtonIndicator chan ButtonIndicator, incomingMessage <-chan QueueOperation, outgoingMessage chan<- QueueOperation, messageSent <-chan QueueOperation, nextDirection chan []int, getNextDirection chan bool,  peersTransmitMSG chan driverState, elevatorStuck <-chan bool) {
	inQueue := 0

	tempQueue = make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}
	queue.queue := QueueMap{queue: make(map[string][][]bool)}
	queue.mux.Lock()
	queue.queue[id] = tempQueue
	queue.mux.Unlock()

	//First is floor, second is direction
	driverStates := DriverStatesMap{states: make(map[string][]int)}
	driverStates.mux.Lock()
	driverStates.states[id] = []int{1, -1}
	driverStates.mux.Unlock()


	for {
		select {
		case outgoingMessage <- updateQueue:
		case peer := <- updateQueueSize:
			updateQueueSize(queue, driverStates, peer)
		case receivedMessageOperation := <- incomingMSG:
			updateQueue(id,receivedMessageOperation,queue,inQueue,getNextDirection)
			updateButtonIndicators(id,receivedMessageOperation,setButtonIndicator)
		case sentMessageOperation := <- messageSent:
			updateQueue(id,sentMessageOperation,queue,inQueue,getNextDirection)
			updateButtonIndicators(id,sentMessageOperation,setButtonIndicator)
		case floor := <- shouldStop:
			shouldStop(id, floor, queue, driverStates, nextDirection)
			driverStates.mux.Lock()
			driverStates.states[id][0] = floor
			peersTransmitMSG <- driverState{id,floor,driverStates.states[id][1]}
			driverStates.mux.Unlock()
		case <-getNextDirection:
			nextDirection(id, queue, driverStates, peersTransmitMSG, nextDirection)
		case <- elevatorStuck:
			driverStates.mux.Lock()
			driverStates.states[id][0] = 100
			peersTransmitMSG <- driverState{id, driverStates[id][0], driverStates[id][1]}
			driverStates.mux.Unlock()
			//redistribute orders
		case calc := <-calcOptimalElevator:
			calculateOptimalElevator(id, driverStates, calc, outgoingMessage)
		}
	}
}

func updateQueue(id int, operation QueueOperation, queue QueueMap, inQueue int, getNextDirection chan<- bool) {
	if operation.isAddOrder {
		queue.mux.Lock()
		queue.queue[operation.elevatorId][operation.floor][operation.button] = operation.isAddOrder
		queue.mux.Unlock()
		inQueue += 1
		if inQueue == 1 {
			getNextDirection <- true
		}
	}
	else {
		for elevID,_ := range queue {
			for i := 0; i < 2; i++  {
				queue.mux.Lock()
				if queue.queue[elevID][operation.floor][i] {
					queue.queue[elevID][operation.floor][i] = operation.isAddOrder
					if elevID == id {
						inQueue -= 1
					}
				}
				queue.mux.Unlock()
			}
		}
		queue.mux.Lock()
		if queue.queue[operation.elevatorId][operation.floor][2] {
			queue.queue[operation.elevatorId][operation.floor][2] = operation.isAddOrder
			inQueue -= 1
		}
		queue.mux.Unlock()
	}
}

func updateQueueSize(queue QueueMap, driverStates DriverStatesMap, peer NewOrLostPeer) {
	if peer.isNew{
		tempQueue := make([][]bool, N_FLOORS)
		for i := range tempQueue {
			tempQueue[i] = make([]bool, N_BUTTONS)
		}
		queue.mux.Lock()
		queue.queue[peer.id] = tempQueue
		queue.mux.Unlock()
		driverStates.mux.Lock()
		driverStates.states[peer.id] = []int{1, -1}
		driverStates.mux.Unlock()
	}
	else{
		//Redistribute orders here
		driverStates.mux.Lock()
		delete(driverStates, peer.id)
		driverStates.mux.Unlock()
		queue.mux.Lock()
		delete(queue, peer.id)
		queue.mux.Unlock()
	}
}

func shouldStop(id string, floor int, queue QueueMap, driverStates DriverStatesMap, nextDirection chan<- []int) {
	driverStates.mux.Lock()
	currentDirection := driverStates.states[id][1]
	driverStates.mux.Unlock()
	if currentDirection == -1 {
		queue.mux.Lock()
		if ((queue.queue[id][floor][2])||(queue.queue[id][floor][1])||(floor == 0)) {
			nextDirection <- []int{0,floor}
		}
		queue.mux.Unlock()
	}
	else{
		queue.mux.Lock()
		if ((queue.queue[id][floor][2])||(queue.queue[id][floor][0])||(floor == N_FLOORS-1)) {
			nextDirection <- []int{0,floor}
		}
		queue.mux.Unlock()
	}
}

func updateButtonIndicators(id string, operation QueueOperation, setButtonIndicator chan<- ButtonIndicator) {
	if operation.isAddOrder {
		if (operation.button!=2) {
			setButtonIndicator <- ButtonIndicator{operation.floor, operation.button, 1}
		}
		else{
			if operation.id == id {
				setButtonIndicator <- ButtonIndicator{operation.floor, operation.button, 1}
			}
		}
	}
	else{
		if operation.button != 2 {
			setButtonIndicator <- ButtonIndicator{operation.floor,0,0}
			setButtonIndicator <- ButtonIndicator{operation.floor,1,0}
		}
		else{
			setButtonIndicator <- ButtonIndicator{operation.floor, operation.button, 0}
		}
	}
}

func calculateOptimalElevator(id string, driverStates DriverStatesMap, order Order, outgoingMessage chan<- QueueOperation) {
	var lowestCost int = 1000
	var lowestCostID string
	n_moves := N_FLOORS-1
	if order.button == 2 {
		lowestCostID = id
	}
	else if order.button == 1 {
		for elevID,list := range queue {
			driverStates.mux.Lock()
			curFloor := driverStates.states[elevID][0]
			curDir := driverStates.states[elevID][1]
			driverStates.mux.Unlock()
			tempCost := 0
			if ((curDir == order.button+1)||(curDir == order.button-2)) {
				if (curDir*(order.floor-curfloor) > 0) {
					tempCost = curDir*(order.floor-curfloor)
				}
				else {
					tempCost = 2*n_moves + curDir*(order.floor-curfloor)
				}
			}
			else if ((curDir == order.button)||(curDir == order.button-1)){
				tempCost = (curDir+1)*n_moves - curDir*(curFloor + order.floor)
			}
			else {
				//FEIL!!
			}
			if tempCost < lowestCost {
				lowestCostID = elevID
			}
		}
	}
	outgoingMessage <- QueueOperation{true, lowestCostID, order.floor, order.button}
}

func nextDirection(id string, queue QueueMap, driverStates DriverStatesMap, peersTransmitMSG chan<- driverState, nextDirection chan<- []int) {
	driverStates.mux.Lock()
	currentDirection := driverStates.states[id][1]
	currentFloor := driverStates.states[id][0]
	driverStates.mux.Unlock()
	updateDirection := 5
	if currentDirection == -1 {
		for floor := currentFloor; floor >= 0; i-- {
			for button := 0; i < N_BUTTONS; button++ {
				queue.mux.Lock()
				if queue.queue[id][floor][button] {
					updateDirection = currentDirection
				}
				queue.mux.Unlock()
			}
		}
	}
	else if currentDirection == 1 {
		for floor := currentFloor; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {
				queue.mux.Lock()
				if queue.queue[id][floor][button] {
					updateDirection = currentDirection
				}
				queue.mux.Unlock()
			}
		}
	}
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			queue.mux.Lock()
			if queue.queue[id][floor][button] {
				if floor < currentFloor {
					updateDirection = -1
				}
				else if floor > currentFloor {
					updateDirection = 1
				}
				else if floor == currentFloor {
					updateDirection = 0
				}
			}
			queue.mux.Unlock()
		}
	}
	if updateDirection != 5 {
		if updateDirection != 0 {
			driverStates.mux.Lock()
			driverStates.states[id][1] = updateDirection
			peersTransmitMSG <- driverState{id, driverStates[id][0], driverStates[id][1]}
			driverStates.mux.Unlock()
		}
		nextDirection <- []int{updateDirection, currentFloor}
	}
}
