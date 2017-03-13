package Queue

import (
	. "../Def"
	"fmt"
	"time"
)

var n_elevators = 1

var inQueue = 0

func RunQueue(id string, initFloor int, chQN ChannelsQueueNetwork, chQF ChannelsQueueFSM, chQD ChannelsQueueDriver) {

	tempQueue := make([][]bool, N_FLOORS)
	for i := range tempQueue {
		tempQueue[i] = make([]bool, N_BUTTONS)
	}

	queue := QueueMap{Queue: make(map[string][][]bool)}

	queue.Mux.Lock()
	queue.Queue[id] = tempQueue
	queue.Mux.Unlock()

	driverStates := DriverStatesMap{States: make(map[string][]int)}

	driverStates.Mux.Lock()
	tempDriverStateList := []int{initFloor, MOTOR_DOWN, IS_STOPPED}
	driverStates.States[id] = tempDriverStateList
	chQN.OutgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
	driverStates.Mux.Unlock()

	removeOrderRedistributedCh := make(chan QueueOperation, 1)
	recalculateCostCh := make(chan Order, 1)
	redistributeOrdersForIdCh := make(chan Redistribute, 1)

	for {

		select {
		case floor := <-chQF.ShouldStopCh:
			shouldStop(id, floor, queue, driverStates, chQF.NextDirectionCh)
			driverStates.Mux.Lock()
			driverStates.States[id][0] = floor
			chQN.OutgoingDriverStateUpdateCh <- DriverState{id, floor, driverStates.States[id][1], driverStates.States[id][2]}
			driverStates.Mux.Unlock()

		case queueUpdateToSend := <-chQD.UpdateQueueCh:
			chQN.OutgoingQueueUpdateCh <- queueUpdateToSend

		case receivedMessageOperation := <-chQN.IncomingQueueUpdateCh:
			updateQueue(id, receivedMessageOperation, queue, chQF.GetNextDirectionCh, chQD.IsDoorOpenCh, chQD.IsDoorOpenResponseCh)
			updateButtonIndicators(id, receivedMessageOperation, chQD.SetButtonIndicatorCh)

		case sentMessageOperation := <-chQN.MessageSentCh:
			updateQueue(id, sentMessageOperation, queue, chQF.GetNextDirectionCh, chQD.IsDoorOpenCh, chQD.IsDoorOpenResponseCh)
			updateButtonIndicators(id, sentMessageOperation, chQD.SetButtonIndicatorCh)

		case orderToCalculate := <-recalculateCostCh:
			calculateOptimalElevator(id, queue, driverStates, orderToCalculate, chQN.OutgoingQueueUpdateCh)

		case messageToSend := <-removeOrderRedistributedCh:
			chQN.OutgoingQueueUpdateCh <- messageToSend

		case <-chQF.GetNextDirectionCh:
			nextDirection(id, queue, driverStates, chQN.OutgoingDriverStateUpdateCh, chQF.NextDirectionCh)

		case <-chQF.ElevatorStuckRedistributeQueueCh:
			driverStates.Mux.Lock()
			driverStates.States[id][0] = 100 //Floor is set to 100
			chQF.NextDirectionCh <- []int{MOTOR_IDLE, 1}
			chQN.OutgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
			driverStates.Mux.Unlock()
			redistributeOrdersForIdCh <- Redistribute{id, false}

			go func() {
				time.Sleep(10 * time.Second)
				chQF.NextDirectionCh <- []int{MOTOR_DOWN, driverStates.States[id][0]}
				driverStates.States[id][1] = MOTOR_DOWN
				driverStates.States[id][2] = IS_MOVING
			}()

		case redistribute := <-redistributeOrdersForIdCh:
			go redistributeOrders(id, redistribute, queue, recalculateCostCh, removeOrderRedistributedCh)

		case peer := <-chQN.UpdateQueueSizeCh:
			updateQueueSize(id, queue, driverStates, peer, redistributeOrdersForIdCh)
			chQN.OutgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1], driverStates.States[id][2]}
			if peer.IsNew {
				go sendCurrentQueue(queue, peer.ElevatorId, chQN.OutgoingQueueUpdateCh)
			}

		case orderToCalc := <-chQD.CalcOptimalElevatorCh:
			calculateOptimalElevator(id, queue, driverStates, orderToCalc, chQN.OutgoingQueueUpdateCh)

		case updatedDriverState := <-chQN.IncomingDriverStateUpdateCh:
			updateDriverState(driverStates, updatedDriverState)
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
			for b := 0; b < 2; b++ {
				queue.Mux.Lock()
				if queue.Queue[elevID][operation.Floor][b] {
					queue.Queue[elevID][operation.Floor][b] = operation.IsAddOrder
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
		fmt.Printf("Queue for id: %s\n", elevID)
		for f := 0; f < N_FLOORS; f++ {
			fmt.Printf("%t\n", queue.Queue[elevID][f])
		}
	}*/
}

func updateQueueSize(id string, queue QueueMap, driverStates DriverStatesMap, peer NewOrLostPeer, redistributeOrdersForIdCh chan<- Redistribute) {
	alreadyInQueue := false
	if peer.IsNew {
		for elevID := range queue.Queue {
			if elevID == peer.ElevatorId {
				alreadyInQueue = true
			}
		}
		if !alreadyInQueue {
			tempQueue := make([][]bool, N_FLOORS)
			for i := range tempQueue {
				tempQueue[i] = make([]bool, N_BUTTONS)
			}
			queue.Mux.Lock()
			queue.Queue[peer.ElevatorId] = tempQueue
			queue.Mux.Unlock()
		}

	} else {
		if peer.ElevatorId != id {
			driverStates.Mux.Lock()
			delete(driverStates.States, peer.ElevatorId)
			driverStates.Mux.Unlock()
			redistributeOrdersForIdCh <- Redistribute{peer.ElevatorId, true}
		}
	}
}

func sendCurrentQueue(queue QueueMap, newPeerId string, outgoingQueueUpdateCh chan<- QueueOperation) {
	fmt.Println("\n")
	queue.Mux.Lock()
	for elevID := range queue.Queue {
		for floor := 0; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {
				if queue.Queue[elevID][floor][button] {
					outgoingQueueUpdateCh <- QueueOperation{true, elevID, floor, button}
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
				tempCost = curDir * (order.Floor - curFloor)
			} else if curDir*(order.Floor-curFloor) < 0 {
				tempCost = curDir * (curFloor - order.Floor)
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
						tempCost += 2
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
	updated := false
	updateDirection := 0
	orderUnder := false
	orderOver := false

	if currentFloor == 100 {
		return
	}

	for button := 0; button < N_BUTTONS; button++ {
		queue.Mux.Lock()
		if queue.Queue[id][currentFloor][button] {
			updateDirection = MOTOR_IDLE
			nextDirectionCh <- []int{updateDirection, currentFloor}
			driverStates.States[id][2] = IS_STOPPED
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
					updated = true
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
						updated = true
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
					updated = true
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
						updated = true
					}
					queue.Mux.Unlock()
				}
			}
		}
	}

	if updated {
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
			if queue.Queue[redistribute.ElevatorId][floor][button] {
				recalculateCostCh <- Order{floor, button}
			}
			queue.Mux.Unlock()
		}
	}
	time.Sleep(20 * time.Millisecond)
	/*if redistribute.ShouldDelete && (id != redistribute.ElevatorId) {
		queue.Mux.Lock()
		delete(queue.Queue, redistribute.ElevatorId)
		queue.Mux.Unlock()
	}*/
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < 2; button++ {
			queue.Mux.Lock()
			queue.Queue[redistribute.ElevatorId][floor][button] = false
			queue.Mux.Unlock()
		}
	}
}

func updateDriverState(driverStates DriverStatesMap, updatedDriverState DriverState) {
	isKeyInDriverstates := false

	for key := range driverStates.States {
		if updatedDriverState.ElevatorId == key {
			isKeyInDriverstates = true
		}
	}

	if !isKeyInDriverstates {
		driverStates.Mux.Lock()
		driverStates.States[updatedDriverState.ElevatorId] = []int{updatedDriverState.LastFloor, updatedDriverState.Direction, updatedDriverState.IsStopped}
		driverStates.Mux.Unlock()

	} else {
		driverStates.Mux.Lock()
		driverStates.States[updatedDriverState.ElevatorId][0] = updatedDriverState.LastFloor
		driverStates.States[updatedDriverState.ElevatorId][1] = updatedDriverState.Direction
		driverStates.States[updatedDriverState.ElevatorId][2] = updatedDriverState.IsStopped
		driverStates.Mux.Unlock()
	}
}
