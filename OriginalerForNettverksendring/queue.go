package Queue

import (
	. "../Def"
	"fmt"
	//"os"
	//"time"
)

var n_elevators = 1

var inQueue = 0

func RunQueue(id string, initFloor int, calcOptimalElevatorCh <-chan Order, updatePeersOnQueueCh <-chan DriverState, updateQueueCh chan QueueOperation, updateQueueSizeCh <-chan NewOrLostPeer, shouldStopCh <-chan int, setButtonIndicatorCh chan ButtonIndicator, incomingMessageCh <-chan QueueOperation, outgoingMessageCh chan<- QueueOperation, messageSentCh <-chan QueueOperation, nextDirectionCh chan []int, getNextDirectionCh chan bool, peersTransmitMessageCh chan DriverState, elevatorStuckCh <-chan bool) {

	tempQueue := make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}
	queue := QueueMap{Queue: make(map[string][][]bool)}
	queue.Mux.Lock()
	queue.Queue[id] = tempQueue
	queue.Mux.Unlock()

	//First is floor, second is direction
	driverStates := DriverStatesMap{States: make(map[string][]int)}
	driverStates.Mux.Lock()
	driverStates.States[id] = []int{initFloor, MOTOR_DOWN}
	driverStates.Mux.Unlock()

	for {
		select {
		case floor := <-shouldStopCh:
			shouldStop(id, floor, queue, driverStates, nextDirectionCh)
			driverStates.Mux.Lock()
			driverStates.States[id][0] = floor
			peersTransmitMessageCh <- DriverState{id, floor, driverStates.States[id][1]}
			driverStates.Mux.Unlock()
		case queueUpdateToSend := <-updateQueueCh:
			fmt.Println("1\n")
			outgoingMessageCh <- queueUpdateToSend
		case peer := <-updateQueueSizeCh:
			fmt.Println("2\n")
			updateQueueSize(queue, driverStates, peer)
		case receivedMessageOperation := <-incomingMessageCh:
			updateQueue(id, receivedMessageOperation, queue, getNextDirectionCh)
			updateButtonIndicators(id, receivedMessageOperation, setButtonIndicatorCh)
		case sentMessageOperation := <-messageSentCh:
			fmt.Println("4\n")
			updateQueue(id, sentMessageOperation, queue, getNextDirectionCh)
			updateButtonIndicators(id, sentMessageOperation, setButtonIndicatorCh)
		case <-getNextDirectionCh:
			fmt.Println("5\n")
			nextDirection(id, queue, driverStates, peersTransmitMessageCh, nextDirectionCh)
		case <-elevatorStuckCh:
			fmt.Println("6\n")
			driverStates.Mux.Lock()
			driverStates.States[id][0] = 100
			peersTransmitMessageCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1]}
			driverStates.Mux.Unlock()
			//redistribute orders
		case calc := <-calcOptimalElevatorCh:
			fmt.Println("7\n")
			calculateOptimalElevator(id, driverStates, calc, outgoingMessageCh)
		case updatedDriverState := <-updatePeersOnQueueCh:
			//fmt.Println("8\n")
			driverStates.Mux.Lock()
			driverStates.States[id][0] = updatedDriverState.LastFloor
			driverStates.States[id][1] = updatedDriverState.Direction
			driverStates.Mux.Unlock()
		}
	}
}

func updateQueue(id string, operation QueueOperation, queue QueueMap, getNextDirectionCh chan<- bool) {
	if operation.IsAddOrder {
		queue.Mux.Lock()
		queue.Queue[operation.ElevatorId][operation.Floor][operation.Button] = operation.IsAddOrder
		queue.Mux.Unlock()
		inQueue++
		fmt.Printf("antallet i køen er: %d\n", inQueue)
		if inQueue == 1 {
			getNextDirectionCh <- true
			fmt.Println("Ingen i kø, så utfører order\n")
		}
	} else {
		for elevID, _ := range queue.Queue {
			for i := 0; i < 2; i++ {
				queue.Mux.Lock()
				if queue.Queue[elevID][operation.Floor][i] {
					queue.Queue[elevID][operation.Floor][i] = operation.IsAddOrder
					if elevID == id {
						inQueue--

					}
				}
				queue.Mux.Unlock()
			}
		}
		queue.Mux.Lock()
		if queue.Queue[operation.ElevatorId][operation.Floor][2] {
			queue.Queue[operation.ElevatorId][operation.Floor][2] = operation.IsAddOrder
			inQueue--
		}
		queue.Mux.Unlock()
	}
}

func updateQueueSize(queue QueueMap, driverStates DriverStatesMap, peer NewOrLostPeer) {
	if peer.IsNew {
		tempQueue := make([][]bool, N_FLOORS)
		for i := range tempQueue {
			tempQueue[i] = make([]bool, N_BUTTONS)
		}
		queue.Mux.Lock()
		queue.Queue[peer.Id] = tempQueue
		queue.Mux.Unlock()
		driverStates.Mux.Lock()
		driverStates.States[peer.Id] = []int{1, -1}
		driverStates.Mux.Unlock()
	} else {
		//Redistribute orders here
		driverStates.Mux.Lock()
		delete(driverStates.States, peer.Id)
		driverStates.Mux.Unlock()
		queue.Mux.Lock()
		delete(queue.Queue, peer.Id)
		queue.Mux.Unlock()
	}
}

func shouldStop(id string, floor int, queue QueueMap, driverStates DriverStatesMap, nextDirectionCh chan<- []int) {
	driverStates.Mux.Lock()
	currentDirection := driverStates.States[id][1]
	driverStates.Mux.Unlock()
	if currentDirection == -1 {
		queue.Mux.Lock()
		if (queue.Queue[id][floor][2]) || (queue.Queue[id][floor][1]) || (floor == 0) {
			nextDirectionCh <- []int{0, floor}
		} else {
			for f := floor - 1; f >= 0; f-- {
				for b := 0; b < N_BUTTONS; b++ {
					if queue.Queue[id][f][b] {
						queue.Mux.Unlock()
						return
					}
				}
			}
			nextDirectionCh <- []int{0, floor}
			return
		}
		queue.Mux.Unlock()
	} else {
		queue.Mux.Lock()
		if (queue.Queue[id][floor][2]) || (queue.Queue[id][floor][0]) || (floor == N_FLOORS-1) {
			nextDirectionCh <- []int{0, floor}
		} else {
			for f := floor + 1; f < N_FLOORS; f++ {
				for b := 0; b < N_BUTTONS; b++ {
					if queue.Queue[id][f][b] {
						queue.Mux.Unlock()
						return
					}
				}
			}
			nextDirectionCh <- []int{0, floor}
		}
		queue.Mux.Unlock()
	}
}

func updateButtonIndicators(id string, operation QueueOperation, setButtonIndicatorCh chan<- ButtonIndicator) {
	if operation.IsAddOrder {
		if operation.Button != 2 {
			setButtonIndicatorCh <- ButtonIndicator{operation.Floor, operation.Button, 1}
		} else {
			if operation.ElevatorId == id {
				setButtonIndicatorCh <- ButtonIndicator{operation.Floor, operation.Button, 1}
			}
		}
	} else {
		setButtonIndicatorCh <- ButtonIndicator{operation.Floor, 0, 0}
		setButtonIndicatorCh <- ButtonIndicator{operation.Floor, 1, 0}
		if operation.ElevatorId == id {
			setButtonIndicatorCh <- ButtonIndicator{operation.Floor, 2, 0}
		}
	}
}

func calculateOptimalElevator(id string, driverStates DriverStatesMap, order Order, outgoingMessageCh chan<- QueueOperation) {
	var lowestCost int = 1000
	var lowestCostID string
	n_moves := N_FLOORS - 1
	if order.Button == 2 {
		lowestCostID = id
	} else {
		for elevID, _ := range driverStates.States {
			driverStates.Mux.Lock()
			curFloor := driverStates.States[elevID][0]
			curDir := driverStates.States[elevID][1]
			driverStates.Mux.Unlock()
			tempCost := 0
			if (curDir == order.Button+1) || (curDir == order.Button-2) {
				if curDir*(order.Floor-curFloor) > 0 {
					tempCost = curDir * (order.Floor - curFloor)
				} else {
					tempCost = 2*n_moves + curDir*(order.Floor-curFloor)
				}
			} else if (curDir == order.Button) || (curDir == order.Button-1) {
				tempCost = (curDir+1)*n_moves - curDir*(curFloor+order.Floor)
			} else {
				//FEIL!!
			}
			if tempCost < lowestCost {
				lowestCostID = elevID
			}
		}
	}
	outgoingMessageCh <- QueueOperation{true, lowestCostID, order.Floor, order.Button}
}

func nextDirection(id string, queue QueueMap, driverStates DriverStatesMap, peersTransmitMessageCh chan<- DriverState, nextDirectionCh chan<- []int) {
	driverStates.Mux.Lock()
	currentDirection := driverStates.States[id][1]
	currentFloor := driverStates.States[id][0]
	driverStates.Mux.Unlock()
	updateDirection := 5
	if currentDirection == -1 {
		for floor := currentFloor; floor >= 0; floor-- {
			for button := 0; button < N_BUTTONS; button++ {
				queue.Mux.Lock()
				if queue.Queue[id][floor][button] {
					updateDirection = currentDirection
				}
				queue.Mux.Unlock()
			}
		}
	} else if currentDirection == 1 {
		for floor := currentFloor; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {
				queue.Mux.Lock()
				if queue.Queue[id][floor][button] {
					updateDirection = currentDirection
				}
				queue.Mux.Unlock()
			}
		}
	}
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			queue.Mux.Lock()
			if queue.Queue[id][floor][button] {
				if floor < currentFloor {
					updateDirection = -1
				} else if floor > currentFloor {
					updateDirection = 1
				} else if floor == currentFloor {
					updateDirection = 0
				}
			}
			queue.Mux.Unlock()
		}
	}
	if updateDirection != 5 {
		if updateDirection != 0 {
			driverStates.Mux.Lock()
			driverStates.States[id][1] = updateDirection
			peersTransmitMessageCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1]}
			driverStates.Mux.Unlock()
		}
		nextDirectionCh <- []int{updateDirection, currentFloor}
	}
}
