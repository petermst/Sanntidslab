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
	operation 	bool
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



func RunQueue(id string, updateQueue <-chan QueueOperation, updateQueueSize <-chan NewOrLostPeer, shouldStop <-chan int, setButtonIndicator chan ButtonIndicator, outgoingMessage chan<- QueueOperation) {

	tempQueue = make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}

	queue := QueueMap{queue: make(map[string][][]bool)}
	queue[id] = tempQueue

	//First is floor, second is direction
	driverStates := make(map[string][]int)
	driverStates[id] = [2]int{1, 0}


	for {
		select {
		case operation <- updateQueue:
			updateQueue(operation, queue)
			outgoingMessage <- operation
		case peer <- updateQueueSize:
			updateQueueSize(queue, peer)
		case receivedMessage <- incomingMSG:
			updateQueue(operation, receivedMessage)
			
		case Messagesent <- :

		}
	}
}

func incomingMSG() {

}

func updateQueue(operation QueueOperation, queue map[string][][]bool) {
	if operation.operation {
		queue[operation.elevatorId][operation.floor][operation.button] = operation.operation
	} 
	else {
		for i := range queue[operation.elevatorId][operation.floor] {
			queue[operation.elevatorId][operation.floor][i] = operation.operation
		}
	}
}

func updateQueueSize(queue map[string][][]bool, peer NewOrLostPeer) {

	if peer.isNew{
		tempQueue = make([][]bool, N_FLOORS)
		for i := range tempQueue {
			tempQueue[i] = make([]bool, N_BUTTONS)
		}
		queue[peer.id] = tempQueue
	}
	else{

		//Redistribute orders here

		delete(queue,peer.id)
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

func updateButtonIndicators(operation QueueOperation, setButtonIndicator chan<- ButtonIndicator) {
	if operation.operation {
		setButtonIndicator <- ButtonIndicator{operation.floor, operation.button, 1}
	}
	else{
		for i := 0; i < 2; i++{
			setButtonIndicator <- ButtonIndicator{operation.floor, i, 0}
		}
	}
}

func calculateElevator() {

}

func nextDirection() {

}

func backupQueue() {

}

func restoreBackup() {

}
