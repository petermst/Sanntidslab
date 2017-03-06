package Queue

import (
	. "../Driver"
	"fmt"
	//"os"
	"time"
)

var n_elevators := 1

type QueueOperation struct {
	operation bool
	elevator int
	floor     int
}

type Order struct {
	floor  int
	button int
}

func RunQueue() {

	queue := initializeQueue()

	for {
		select {
		case operation <- updateQueue:
			updateQueue(queue, operation)
		}
	}
}

func initializeQueue() ([][][]*bool){
	
	queue := make([][][]*bool,n_elevators)

	for i := n_elevators{
		queue[i] = make([][]*bool,N_FLOORS)
		for j := range queue[i]{
			queue[i][j] = make([]*bool,N_BUTTONS)
			for k :=  range queue[i][j]{
				queue[i][j][k] = false
			}
		}
	}
	return queue
}

func incomingMSG() {

}

func updateQueue(queue [][][]*bool, operation QueueOperation) {
	for i := range queue[][]	
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
