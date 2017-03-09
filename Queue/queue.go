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

func RunQueue(id string,calcOptimalElevator <-chan Order, updateQueue <-chan QueueOperation, updateQueueSize <-chan NewOrLostPeer, shouldStop <-chan int, setButtonIndicator chan ButtonIndicator, incomingMessage <-chan QueueOperation, outgoingMessage chan<- QueueOperation, messageSent <-chan QueueOperation, nextDirection chan []int, getNextDirection chan bool,  peersTransmitMSG chan driverState, elevatorStuck <-chan bool) {
	inQueue := 0

	tempQueue = make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}
	queue := QueueMap{queue: make(map[string][][]bool)}
	queue.mux.Lock()
	queue[id] = tempQueue
	queue.mux.Unlock()

	//First is floor, second is direction
	driverStates := make(map[string][]int)
	driverStates[id] = []int{1, -1}


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
		case <-getNextDirection:
			nextDirection(id, queue, driverStates, peersTransmitMSG, nextDirection)
		case <- elevatorStuck:
			driverStates[id][0] = 100
			peersTransmitMSG <- driverState{id, driverStates[id][0], driverStates[id][1]}
			//redistribute orders
		case calc := <-calcOptimalElevator:
			calculateOptimalElevator(id, driverStates, calc, outgoingMessage)
		}
	}
}

func updateQueue(id int, operation QueueOperation, queue QueueMap, inQueue int, getNextDirection chan<- bool) {
	if operation.isAddOrder {
		queue.mux.Lock()
		queue[operation.elevatorId][operation.floor][operation.button] = operation.isAddOrder
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
				defer queue.mux.Unlock()
				if queue[elevID][operation.floor][i] {
					queue.mux.Lock()
					queue[elevID][operation.floor][i] = operation.isAddOrder
					queue.mux.Unlock()
					if elevID == id {
						inQueue -= 1
					}
				}
			}
		}
		queue.mux.Lock()
		defer queue.mux.Unlock()
		if queue[operation.elevatorId][operation.floor][2] {
			queue.mux.Lock()
			queue[operation.elevatorId][operation.floor][2] = operation.isAddOrder
			queue.mux.Unlock()
			inQueue -= 1
		}
	}
}

func updateQueueSize(queue QueueMap, driverStates map[string][]int, peer NewOrLostPeer) {
	if peer.isNew{
		tempQueue := make([][]bool, N_FLOORS)
		for i := range tempQueue {
			tempQueue[i] = make([]bool, N_BUTTONS)
		}
		queue.mux.Lock()
		queue[peer.id] = tempQueue
		queue.mux.Unlock()
		driverStates[peer.id] = []int{1, -1}
	}
	else{
		//Redistribute orders here
		delete(driverStates, peer.id)
		delete(queue, peer.id)
	}
}

func shouldStop(id string, floor int, queue QueueMap, driverStates map[string][]int, nextDirection chan<- []int) {
	currentDirection := driverStates[id][1]
	if currentDirection == -1 {
		queue.mux.Lock()
		defer queue.mux.Unlock()
		if ((queue[id][floor][2])||(queue[id][floor][1])||(floor == 0)) {
			nextDirection <- []int{0,floor}
		}
	}
	else{
		queue.mux.Lock()
		defer queue.mux.Unlock()
		if ((queue[id][floor][2])||(queue[id][floor][0])||(floor == N_FLOORS-1)) {
			nextDirection <- []int{0,floor}
		}
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

func calculateOptimalElevator(id string, elevatorStates map[string][]int, order Order, outgoingMessage chan<- QueueOperation) {
	var lowestCost int = 1000
	var lowestCostID string
	n_moves := N_FLOORS-1
	if order.button == 2 {
		lowestCostID = id
	}
	else if order.button == 1 {
		for elevID,list := range queue {
			curFloor := driverStates[elevID][0]
			curDir := driverStates[elevID][1]
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

func nextDirection(id string, queue QueueMap, driverStates map[string][]int, peersTransmitMSG chan<- driverState, nextDirection chan<- []int) {
	currentDirection := driverStates[id][1]
	currentFloor := driverStates[id][0]
	updateDirection := 5
	if currentDirection == -1 {
		for floor := currentFloor; floor >= 0; i-- {
			for button := 0; i < N_BUTTONS; button++ {
				queue.mux.Lock()
				defer queue.mux.Unlock()
				if queue[id][floor][button] {
					updateDirection = currentDirection
				}
			}
		}
	}
	else if currentDirection == 1 {
		for floor := currentFloor; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {
				queue.mux.Lock()
				defer queue.mux.Unlock()
				if queue[id][floor][button] {
					updateDirection = currentDirection
				}
			}
		}
	}
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			queue.mux.Lock()
			defer queue.mux.Unlock()
			if queue[id][floor][button] {
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
		}
	}
	if updateDirection != 5 {
		if updateDirection != 0 {
			driverStates[id][1] = updateDirection
			peersTransmitMSG <- driverState{id, driverStates[id][0], driverStates[id][1]}
		}
		nextDirection <- []int{updateDirection, currentFloor}
	}
}
