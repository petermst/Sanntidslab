package FSM

import (
	. "../Def"
	//"fmt"
	//"os"
	//"time"
)

func RunFSM(id string, setMotorDirectionCh chan int, startDoorTimerCh chan bool, eventElevatorStuckCh <-chan bool, eventAtFloorCh <-chan int, eventDoorTimeoutCh <-chan bool, nextDirectionCh <-chan []int, shouldStopCh chan<- int, getNextDirectionCh chan<- bool, elevatorStuckUpdateQueueCh chan<- bool, updateQueueCh chan<- QueueOperation) {
	state := STATE_IDLE

	for {
		select {
		case <-eventElevatorStuckCh:
			state = STATE_STUCK
			elevatorStuckUpdateQueueCh <- true
		case floor := <-eventAtFloorCh:
			eventAtFloor(state, floor, shouldStopCh)
		case directionAndFloor := <-nextDirectionCh:
			state = eventNewDirection(id, state, directionAndFloor, startDoorTimerCh, updateQueueCh)
			setMotorDirectionCh <- directionAndFloor[0]
		case <-eventDoorTimeoutCh:
			state = eventDoorTimeout(state)
			getNextDirectionCh <- true
		}
	}
}

func eventAtFloor(state State, floor int, shouldStopCh chan<- int) State {
	switch state {
	case STATE_MOVING:
		shouldStopCh <- floor
	default:
		return state
	}
	return state
}

func eventDoorTimeout(state State) State {
	switch state {
	case STATE_DOOR_OPEN:

		return STATE_IDLE
	default:
		return state
	}
}

func eventNewDirection(id string, state State, directionAndFloor []int, startDoorTimerCh chan<- bool, updateQueueCh chan<- QueueOperation) State {
	switch state {
	case STATE_MOVING:
	case STATE_IDLE:
		if directionAndFloor[0] == 0 {
			startDoorTimerCh <- true
			QueOpe := QueueOperation{false, id, directionAndFloor[1], 0}
			updateQueueCh <- QueOpe
			return STATE_DOOR_OPEN
		} else {
			return STATE_MOVING
		}
	default:
		return state
	}
	return state
}
