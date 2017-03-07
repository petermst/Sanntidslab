package Queue

import (
	. "../Driver"
	"fmt"
	//"os"
	"time"
)

var n_elevators = 1

type QueueOperation struct {
	operation bool
	elevator  int
	floor     int
}

type Order struct {
	floor  int
	button int
}

type QueueMap struct {
	mux   sync.Mutex
	queue map[string][][]bool
}

func RunQueue(id string) {
	tempQueue = make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}

	queue := QueueMap{queue: make(map[string][][]bool)}
	queue[id] = tempQueue

	driverstates := make(map[string][]int)
	driverstates[id] = [2]int{1, 0}

	for {
		select {
		case operation <- updateQueue:
			updateQueue(queue, operation)
		}
	}
}

func incomingMSG() {

}

func updateQueue(operation QueueOperation) {

}

func updateQueueSize() {

}

func shouldStop() {

}

func calculateElevator() {

}

func nextDirection() {

}

func backupQueue() {

}

func restoreBackup() {

}
