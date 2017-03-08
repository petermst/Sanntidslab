package Queue

import (
	. "../Driver"
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



func RunQueue(id string, updateQueue <-chan QueueOperation, updateQueueSize <-chan NewOrLostPeer, shouldStop <-chan int, setButtonIndicator chan ButtonIndicator, incomingMessage <-chan QueueOperation, outgoingMessage chan<- QueueOperation, messageSent <-chan QueueOperation, nextDirection chan []int, getNextDirection <-chan bool) {

	tempQueue = make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}

	queue := QueueMap{queue: make(map[string][][]bool)}
	queue[id] = tempQueue

	//First is floor, second is direction
	driverStates := make(map[string][]int)
	driverStates[id] = [2]int{1, -1}


	for {
		select {
		case outgoingMessage <- updateQueue:
		case peer := <- updateQueueSize:
			updateQueueSize(queue, driverStates, peer)
		case receivedMessageOperation := <- incomingMSG:
			updateQueue(receivedMessageOperation, queue)
			updateButtonIndicators(id,receivedMessageOperation,setButtonIndicator)
		case sentMessageOperation := <- messageSent:
			updateQueue(sentMessageOperation,queue)
			updateButtonIndicators(id,sentMessageOperation,setButtonIndicator)
		case floor := <- shouldStop:
			shouldStop(id, floor, queue, driverStates, nextDirection)
		case <-getNextDirection:
			nextDirection()
		}
	}
}

func updateQueue(operation QueueOperation, queue map[string][][]bool) {
	if operation.isAddOrder {
		queue[operation.elevatorId][operation.floor][operation.button] = operation.isAddOrder
	}
	else {
		for i := range queue[operation.elevatorId][operation.floor] {
			queue[operation.elevatorId][operation.floor][i] = operation.isAddOrder
		}
	}
}

func updateQueueSize(queue map[string][][]bool, driverStates map[string][]int, peer NewOrLostPeer) {

	if peer.isNew{
		tempQueue := make([][]bool, N_FLOORS)
		tempDriveState := make([]int,)
		for i := range tempQueue {
			tempQueue[i] = make([]bool, N_BUTTONS)
		}
		queue[peer.id] = tempQueue
		driverStates[peer.id] = [2]int{1, -1}
	}
	else{

		//Redistribute orders here
		delete(driverStates, peer.id)
		delete(queue, peer.id)
	}
}

func shouldStop(id string,floor int, queue map[string][][]bool, driverStates map[string][]int, nextDirection chan<- []int) {
	currentDirection := driverStates[id][1]
	if (queue[id][floor][2]) || (currentDirection==1 && queue[id][floor][0]) || (currentDirection==-1 && queue[id][floor][1]) {
		nextDirection <- [2]int{0,floor}
	}
	else{
		nextDirection <- [2]int{currentDirection,floor}
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

func calculateOptimalElevator(id string, queue map[string][][]bool, elevatorStates map[string][]int, order Order, outgoingMessage chan<- QueueOperation) {
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
			if ((curDir == order.button+1) {
				if (curDir*(order.floor-curfloor) > 0) {
					tempCost = curDir*(order.floor-curfloor)
				}
				else {
					tempCost = 2*n_moves + curDir*(order.floor-curfloor)
				}
			}
			else if (curDir == order.button-2){
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

func nextDirection(id string, queue map[string][][]bool, elevatorStates map[string][]int) {
	
}

