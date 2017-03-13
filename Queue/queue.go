package Queue

import (
	. "../Def"
	"fmt"
	//"os"
	"time"
)

var n_elevators = 1

var inQueue = 0

func RunQueue(id string, initFloor int, calcOptimalElevatorCh <-chan Order, updateQueueCh chan QueueOperation, updateQueueSizeCh <-chan NewOrLostPeer, shouldStopCh <-chan int, setButtonIndicatorCh chan ButtonIndicator, messageSentCh <-chan QueueOperation, nextDirectionCh chan []int, getNextDirectionCh chan bool, elevatorStuckCh <-chan bool, outgoingQueueUpdateCh chan<- QueueOperation, incomingQueueUpdateCh <-chan QueueOperation, outgoingDriverStateUpdateCh chan<- DriverState, incomingDriverStateUpdateCh <-chan DriverState, isDoorOpenCh chan<- bool, isDoorOpenResponseCh <-chan bool) {

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
	driverStateList := []int{initFloor, MOTOR_DOWN, IS_STOPPED}
	driverStates.States[id] = driverStateList
	outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
	driverStates.Mux.Unlock()

	removeOrderRedistributedCh := make(chan QueueOperation, 1)
	recalculateCostCh := make(chan Order, 1)
	redistributeOrdersForIdCh := make(chan Redistribute, 1)

	isKeyInDriverstates := false

	for {
		select {
		case floor := <-shouldStopCh:
			shouldStop(id, floor, queue, driverStates, nextDirectionCh)
			driverStates.Mux.Lock()
			fmt.Printf("Setter driverstates til floor: %d\n", floor)
			driverStates.States[id][0] = floor
			outgoingDriverStateUpdateCh <- DriverState{id, floor, driverStates.States[id][1], driverStates.States[id][2]}
			driverStates.Mux.Unlock()
		case queueUpdateToSend := <-updateQueueCh:
			outgoingQueueUpdateCh <- queueUpdateToSend

		case receivedMessageOperation := <-incomingQueueUpdateCh:
			updateQueue(id, receivedMessageOperation, queue, getNextDirectionCh, isDoorOpenCh, isDoorOpenResponseCh)
			updateButtonIndicators(id, receivedMessageOperation, setButtonIndicatorCh)
		case sentMessageOperation := <-messageSentCh:
			updateQueue(id, sentMessageOperation, queue, getNextDirectionCh, isDoorOpenCh, isDoorOpenResponseCh)
			updateButtonIndicators(id, sentMessageOperation, setButtonIndicatorCh)
		case newOrderToCalculate := <-recalculateCostCh:
			calculateOptimalElevator(id, queue, driverStates, newOrderToCalculate, outgoingQueueUpdateCh)
		case messageToSend := <-removeOrderRedistributedCh:
			outgoingQueueUpdateCh <- messageToSend
		case <-getNextDirectionCh:
			nextDirection(id, queue, driverStates, outgoingDriverStateUpdateCh, nextDirectionCh)
		case redistribute := <-redistributeOrdersForIdCh:
			go redistributeOrders(id, redistribute, queue, recalculateCostCh, removeOrderRedistributedCh)
		case <-elevatorStuckCh:
			driverStates.Mux.Lock()
			driverStates.States[id][0] = 100
			nextDirectionCh <- []int{MOTOR_IDLE, 1}
			outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
			driverStates.Mux.Unlock()
			redistributeOrdersForIdCh <- Redistribute{id, false}
			go func() {
				time.Sleep(10 * time.Second)
				nextDirectionCh <- []int{MOTOR_DOWN, driverStates.States[id][0]}
				driverStates.States[id][1] = MOTOR_DOWN
				driverStates.States[id][2] = IS_MOVING
			}()
		case peer := <-updateQueueSizeCh:
			updateQueueSize(id, queue, driverStates, peer, redistributeOrdersForIdCh)
			outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
			if peer.IsNew {
				go sendCurrentQueue(queue, peer.Id, outgoingQueueUpdateCh)
			}
		case calc := <-calcOptimalElevatorCh:
			calculateOptimalElevator(id, queue, driverStates, calc, outgoingQueueUpdateCh)
		case updatedDriverState := <-incomingDriverStateUpdateCh:
			for key := range driverStates.States {
				if updatedDriverState.Id == key {
					isKeyInDriverstates = true
				}
			}
			if !isKeyInDriverstates {
				driverStates.Mux.Lock()
				driverStates.States[updatedDriverState.Id] = []int{updatedDriverState.LastFloor, updatedDriverState.Direction, updatedDriverState.IsStopped}
				driverStates.Mux.Unlock()
			} else {
				driverStates.Mux.Lock()
				driverStates.States[updatedDriverState.Id][0] = updatedDriverState.LastFloor
				driverStates.States[updatedDriverState.Id][1] = updatedDriverState.Direction
				driverStates.States[updatedDriverState.Id][2] = updatedDriverState.IsStopped
				driverStates.Mux.Unlock()
				isKeyInDriverstates = false
			}
		}
	}
}

func updateQueue(id string, operation QueueOperation, queue QueueMap, getNextDirectionCh chan<- bool, isDoorOpenCh chan<- bool, isDoorOpenResponseCh <-chan bool) {
	if operation.IsAddOrder {
		queue.Mux.Lock()
		if !queue.Queue[operation.ElevatorId][operation.Floor][operation.Button] {
			queue.Queue[operation.ElevatorId][operation.Floor][operation.Button] = operation.IsAddOrder

			if operation.ElevatorId == id {
				inQueue++
				if inQueue == 1 {
					isDoorOpenCh <- true
					doorOpen := <-isDoorOpenResponseCh
					if !doorOpen {
						getNextDirectionCh <- true
					}
				}
			}
		}
		queue.Mux.Unlock()
	} else {
		for elevID := range queue.Queue {
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
			if operation.ElevatorId == id {
				inQueue--
			}
		}
		queue.Mux.Unlock()
	}
	/*for elevID := range queue.Queue {
		for floor := N_FLOORS - 1; floor >= 0; floor-- {
			queue.Mux.Lock()
			fmt.Printf("%t\n", queue.Queue[elevID][floor])
			queue.Mux.Unlock()
		}
		fmt.Println("\n")
	}*/
}

func updateQueueSize(id string, queue QueueMap, driverStates DriverStatesMap, peer NewOrLostPeer, redistributeOrdersForIdCh chan<- Redistribute) {
	if peer.IsNew {
		tempQueue := make([][]bool, N_FLOORS)
		for i := range tempQueue {
			tempQueue[i] = make([]bool, N_BUTTONS)
		}
		queue.Mux.Lock()
		queue.Queue[peer.Id] = tempQueue
		queue.Mux.Unlock()

	} else {
		if peer.Id != id {
			driverStates.Mux.Lock()
			delete(driverStates.States, peer.Id)
			driverStates.Mux.Unlock()
			redistributeOrdersForIdCh <- Redistribute{peer.Id, true}
			fmt.Printf("redistributeOrdersForIdCh blir satt med id: %s\n", peer.Id)
			for key := range driverStates.States {
				fmt.Printf("Dette er id i driverStates: %s\n", key)
			}
		}
	}
}

func sendCurrentQueue(queue QueueMap, newPeerId string, outgoingQueueUpdateCh chan<- QueueOperation) {
	queue.Mux.Lock()
	for elevID := range queue.Queue {
		if !(elevID == newPeerId) {
			for floor := 0; floor < N_FLOORS; floor++ {
				for button := 0; button < N_BUTTONS; button++ {
					if queue.Queue[elevID][floor][button] {
						outgoingQueueUpdateCh <- QueueOperation{true, elevID, floor, button}
					}
				}
			}
		}
	}
	queue.Mux.Unlock()
}

func shouldStop(id string, floor int, queue QueueMap, driverStates DriverStatesMap, nextDirectionCh chan<- []int) {
	driverStates.Mux.Lock()
	currentDirection := driverStates.States[id][1]
	driverStates.Mux.Unlock()
	if currentDirection == MOTOR_DOWN {
		queue.Mux.Lock()
		if (queue.Queue[id][floor][2]) || (queue.Queue[id][floor][1]) || (floor == 0) {
			nextDirectionCh <- []int{0, floor}
			driverStates.States[id][2] = IS_STOPPED
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
			driverStates.States[id][2] = IS_STOPPED
		}
		queue.Mux.Unlock()
	} else if currentDirection == MOTOR_UP {
		queue.Mux.Lock()
		if (queue.Queue[id][floor][2]) || (queue.Queue[id][floor][0]) || (floor == N_FLOORS-1) {
			nextDirectionCh <- []int{0, floor}
			driverStates.States[id][2] = IS_STOPPED
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
			driverStates.States[id][2] = IS_STOPPED
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

func calculateOptimalElevator(id string, queue QueueMap, driverStates DriverStatesMap, order Order, outgoingQueueUpdateCh chan<- QueueOperation) {

	lowestCost := 1000
	lowestCostID := ""
	if order.Button == 2 {
		lowestCostID = id
	} else {
		for elevID := range driverStates.States {
			driverStates.Mux.Lock()
			curFloor := driverStates.States[elevID][0]
			curDir := driverStates.States[elevID][1]
			curIsStopped := driverStates.States[elevID][2]
			driverStates.Mux.Unlock()
			tempCost := 0

			if curDir*(order.Floor-curFloor) > 0 {
				tempCost = 2 * curDir * (order.Floor - curFloor)
			} else if curDir*(order.Floor-curFloor) < 0 {
				tempCost = 2 * curDir * (curFloor - order.Floor)
			} else {
				for b := 0; b < N_BUTTONS; b++ {
					queue.Mux.Lock()
					if queue.Queue[elevID][order.Floor][b] {
						tempCost = 0
					}
					queue.Mux.Unlock()
				}
			}
			for f := 0; f < N_FLOORS; f++ {
				for b := 0; b < N_BUTTONS; b++ {
					queue.Mux.Lock()
					if queue.Queue[elevID][f][b] {
						tempCost += 1
					}
					queue.Mux.Unlock()
				}
			}
			if (curIsStopped == 1) && (order.Floor == curFloor) {
				tempCost = 0
			}
			if tempCost == 0 {
				outgoingQueueUpdateCh <- QueueOperation{true, elevID, order.Floor, order.Button}
				return
			} else if tempCost < lowestCost {
				lowestCostID = elevID
				lowestCost = tempCost
			}

		}
	}
	outgoingQueueUpdateCh <- QueueOperation{true, lowestCostID, order.Floor, order.Button}
}

func nextDirection(id string, queue QueueMap, driverStates DriverStatesMap, outgoingDriverStateUpdateCh chan<- DriverState, nextDirectionCh chan<- []int) {
	driverStates.Mux.Lock()
	currentDirection := driverStates.States[id][1]
	currentFloor := driverStates.States[id][0]
	driverStates.Mux.Unlock()
	updateDirection := 5
	orderUnder := false
	orderOver := false

	if currentFloor == 100 {
		return
	}

	fmt.Printf("floor: %d\n", currentFloor)
	for button := 0; button < N_BUTTONS; button++ {
		queue.Mux.Lock()
		if queue.Queue[id][currentFloor][button] {
			updateDirection = 0
			nextDirectionCh <- []int{updateDirection, currentFloor}
			driverStates.States[id][2] = IS_STOPPED
			fmt.Println("Nå bør den bli i samme etasje\n")
			queue.Mux.Unlock()
			return
		}
		queue.Mux.Unlock()
	}

	if currentDirection == MOTOR_DOWN {
		for floor := currentFloor - 1; floor >= 0; floor-- {
			for button := 0; button < N_BUTTONS; button++ {
				queue.Mux.Lock()
				if queue.Queue[id][floor][button] {
					updateDirection = MOTOR_DOWN
					orderUnder = true
				}
				queue.Mux.Unlock()
			}
		}
		if !orderUnder {
			for floor := currentFloor; floor < N_FLOORS; floor++ {
				for button := 0; button < N_BUTTONS; button++ {
					queue.Mux.Lock()
					if queue.Queue[id][floor][button] {
						updateDirection = MOTOR_UP
					}
					queue.Mux.Unlock()
				}
			}
		}
	} else if currentDirection == MOTOR_UP {
		for floor := currentFloor + 1; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {
				queue.Mux.Lock()
				if queue.Queue[id][floor][button] {
					updateDirection = MOTOR_UP
					orderOver = true
				}
				queue.Mux.Unlock()
			}
		}
		if !orderOver {
			for floor := currentFloor; floor >= 0; floor-- {
				for button := 0; button < N_BUTTONS; button++ {
					queue.Mux.Lock()
					if queue.Queue[id][floor][button] {
						updateDirection = MOTOR_DOWN
					}
					queue.Mux.Unlock()
				}
			}
		}
	}
	if updateDirection != 5 {
		if updateDirection != 0 {
			driverStates.Mux.Lock()
			driverStates.States[id][1] = updateDirection
			driverStates.States[id][2] = IS_MOVING
			outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
			driverStates.Mux.Unlock()
		}
		nextDirectionCh <- []int{updateDirection, currentFloor}

	}
}

func redistributeOrders(id string, redistribute Redistribute, queue QueueMap, recalculateCostCh chan<- Order, removeOrderRedistributedCh chan<- QueueOperation) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < 2; button++ {
			queue.Mux.Lock()
			if queue.Queue[redistribute.Id][floor][button] {
				recalculateCostCh <- Order{floor, button}
			}
			queue.Mux.Unlock()
		}
	}
	if redistribute.ShouldDelete && (id != redistribute.Id) {
		queue.Mux.Lock()
		delete(queue.Queue, redistribute.Id)
		queue.Mux.Unlock()
	}
}
