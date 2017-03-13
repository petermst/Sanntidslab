package FSM

import (
	. "../Def"
	"fmt"
	//"os"
	//"time"
)

func RunFSM(id string, setMotorDirectionCh chan int, startDoorTimerCh chan bool, eventElevatorStuckCh <-chan bool, eventAtFloorCh <-chan int, nextDirectionCh <-chan []int, shouldStopCh chan int, eventDoorTimeoutCh <-chan bool, getNextDirectionCh chan<- bool, elevatorStuckUpdateQueueCh chan<- bool) {
	state := STATE_IDLE
	for {
		fmt.Printf("State før select-case: %d\n", state)
		select {
		case <-eventElevatorStuckCh:
			state = STATE_STUCK
			elevatorStuckUpdateQueueCh <- true
		case floor := <-eventAtFloorCh:
			eventAtFloor(state, floor, shouldStopCh, getNextDirectionCh)
		case directionAndFloor := <-nextDirectionCh:
			state = eventNewDirection(id, state, directionAndFloor, startDoorTimerCh)
			setMotorDirectionCh <- directionAndFloor[0]
		case <-eventDoorTimeoutCh:
			state = eventDoorTimeout(state)
			getNextDirectionCh <- true
		}
	}
}

func eventAtFloor(state State, floor int, shouldStopCh chan<- int, getNextDirectionCh chan<- bool) State {
	fmt.Printf("eventAtFloor State: %d\n", state)
	switch state {
	case STATE_MOVING:
		shouldStopCh <- floor
		return state
	case STATE_STUCK:
		if floor == 0 {
			shouldStopCh <- floor
		}
		return state
	default:
		return state
	}
}

func eventDoorTimeout(state State) State {
	fmt.Printf("eventDoorTimeout State: %d\n", state)
	switch state {
	case STATE_DOOR_OPEN:
		return STATE_IDLE
	default:
		return state
	}
}

func eventNewDirection(id string, state State, directionAndFloor []int, startDoorTimerCh chan<- bool) State {
	fmt.Printf("eventNewDirection State: %d\n", state)
	switch state {
	case STATE_MOVING:
		fallthrough
	case STATE_IDLE:
		if directionAndFloor[0] == MOTOR_IDLE {
			startDoorTimerCh <- true
			return STATE_DOOR_OPEN
		} else {
			return STATE_MOVING
		}
	case STATE_STUCK:
		if directionAndFloor[0] == MOTOR_IDLE {
			return STATE_STUCK
		} else {
			return STATE_MOVING
		}

	default:
		return state
	}
}
