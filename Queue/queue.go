package Queue

import (
	. "../Def"
	"fmt"
	//"os"
	//"time"
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
	driverStateList := []int{initFloor, MOTOR_DOWN}
	driverStates.States[id] = driverStateList
	outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1]}
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
			driverStates.States[id][0] = floor
			outgoingDriverStateUpdateCh <- DriverState{id, floor, driverStates.States[id][1]}
			driverStates.Mux.Unlock()
		case queueUpdateToSend := <-updateQueueCh:
			outgoingQueueUpdateCh <- queueUpdateToSend
		case peer := <-updateQueueSizeCh:
			updateQueueSize(id, queue, driverStates, peer, redistributeOrdersForIdCh)
			outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1]}
			if peer.IsNew {
				go sendCurrentQueue(queue, peer.Id, outgoingQueueUpdateCh)
			}
		case receivedMessageOperation := <-incomingQueueUpdateCh:
			updateQueue(id, receivedMessageOperation, queue, getNextDirectionCh, isDoorOpenCh, isDoorOpenResponseCh)
			updateButtonIndicators(id, receivedMessageOperation, setButtonIndicatorCh)
		case sentMessageOperation := <-messageSentCh:
			updateQueue(id, sentMessageOperation, queue, getNextDirectionCh, isDoorOpenCh, isDoorOpenResponseCh)
			updateButtonIndicators(id, sentMessageOperation, setButtonIndicatorCh)
		case newOrderToCalculate := <-recalculateCostCh:
			calculateOptimalElevator(id, driverStates, newOrderToCalculate, outgoingQueueUpdateCh)
		case messageToSend := <-removeOrderRedistributedCh:
			outgoingQueueUpdateCh <- messageToSend
		case <-getNextDirectionCh:
			nextDirection(id, queue, driverStates, outgoingDriverStateUpdateCh, nextDirectionCh)
		case redistribute := <-redistributeOrdersForIdCh:
			redistributeOrders(redistribute, queue, recalculateCostCh, removeOrderRedistributedCh)
		case <-elevatorStuckCh:
			driverStates.Mux.Lock()
			driverStates.States[id][0] = 100
			outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1]}
			driverStates.Mux.Unlock()
			redistributeOrdersForIdCh <- Redistribute{id, false}

		case calc := <-calcOptimalElevatorCh:
			calculateOptimalElevator(id, driverStates, calc, outgoingQueueUpdateCh)
		case updatedDriverState := <-incomingDriverStateUpdateCh:
			for key, _ := range driverStates.States {
				if updatedDriverState.Id == key {
					isKeyInDriverstates = true
				}
			}
			if !isKeyInDriverstates {
				driverStates.Mux.Lock()
				driverStates.States[updatedDriverState.Id] = []int{updatedDriverState.LastFloor, updatedDriverState.Direction}
				driverStates.Mux.Unlock()
				isKeyInDriverstates = false
			} else {
				driverStates.Mux.Lock()
				driverStates.States[updatedDriverState.Id][0] = updatedDriverState.LastFloor
				driverStates.States[updatedDriverState.Id][1] = updatedDriverState.Direction
				driverStates.Mux.Unlock()
			}
		}
		fmt.Printf("Driverstates er:\n Floor: %d\n Direction: %d\n", driverStates.States[id][0], driverStates.States[id][1])
	}
}

func updateQueue(id string, operation QueueOperation, queue QueueMap, getNextDirectionCh chan<- bool, isDoorOpenCh chan<- bool, isDoorOpenResponseCh <-chan bool) {
	if operation.IsAddOrder {
		queue.Mux.Lock()
		if !queue.Queue[operation.ElevatorId][operation.Floor][operation.Button] {
			queue.Queue[operation.ElevatorId][operation.Floor][operation.Button] = operation.IsAddOrder

			fmt.Printf("antallet i køen er: %d\n", inQueue)

			if operation.ElevatorId == id {
				inQueue++
				if inQueue == 1 {
					isDoorOpenCh <- true
					doorOpen := <-isDoorOpenResponseCh
					if !doorOpen {
						getNextDirectionCh <- true
						fmt.Println("Ingen i kø, utfører ordre")
					}
				}
			}
		}
		queue.Mux.Unlock()
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
			if operation.ElevatorId == id {
				inQueue--
			}
		}
		queue.Mux.Unlock()
	}
	for elevID, _ := range queue.Queue {
		fmt.Printf("Elevator ID: %s\n", elevID)
		for floor := N_FLOORS - 1; floor >= 0; floor-- {
			queue.Mux.Lock()
			fmt.Printf("%t\n", queue.Queue[elevID][floor])
			queue.Mux.Unlock()
		}
		fmt.Println("\n")
	}
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
		driverStates.Mux.Lock()
		delete(driverStates.States, peer.Id)
		driverStates.Mux.Unlock()
		redistributeOrdersForIdCh <- Redistribute{peer.Id, true}
		

	}
}

func sendCurrentQueue(queue QueueOperation, newPeerId string, outgoingQueueUpdateCh chan<- QueueOperation) {
	queue.Mux.Lock()
	for elevID, _ := range queue.Queue {
		if !(elevID == newPeerId){
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
		}
		queue.Mux.Unlock()
	} else if currentDirection == MOTOR_UP {
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

func calculateOptimalElevator(id string, driverStates DriverStatesMap, order Order, outgoingQueueUpdateCh chan<- QueueOperation) {
	for element, _ := range driverStates.States {
		fmt.Printf("Elevator ID: %s\n", element)
		fmt.Printf("Driverstates: %d\n\n", driverStates.States[element])
	}

	lowestCost := 1000
	lowestCostID := ""
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
				lowestCost = tempCost
			}
		}
	}
	fmt.Printf("Optimal elevator: %s, floor: %d, button: %d\n", lowestCostID, order.Floor, order.Button)
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

	fmt.Printf("floor: %d\n", currentFloor)
	for button := 0; button < N_BUTTONS; button++ {
		queue.Mux.Lock()
		if queue.Queue[id][currentFloor][button] {
			updateDirection = 0
			nextDirectionCh <- []int{updateDirection, currentFloor}
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
			outgoingDriverStateUpdateCh <- DriverState{id, driverStates.States[id][0], driverStates.States[id][1]}
			driverStates.Mux.Unlock()
		}
		nextDirectionCh <- []int{updateDirection, currentFloor}
	}
}

func redistributeOrders(redistribute Redistribute, queue QueueMap, recalculateCostCh chan<- Order, removeOrderRedistributedCh chan<- QueueOperation) {
	var updated bool
	for floor := 0; floor < N_FLOORS; floor++ {
		updated = false
		for button := 0; button < 2; button++ {
			queue.Mux.Lock()
			if queue.Queue[redistribute.Id][floor][button] {
				recalculateCostCh <- Order{floor, button}
				updated = true
			}
		}
		if updated {
			removeOrderRedistributedCh <- QueueOperation{false, redistribute.Id, floor, 0}
		}
	}
	if redistribute.ShouldDelete {
		queue.Mux.Lock()
		delete(queue.Queue, redistribute.Id)
		queue.Mux.Unlock()
	}
}
