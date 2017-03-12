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
		select {
		case <-eventElevatorStuckCh:
			state = STATE_STUCK
			elevatorStuckUpdateQueueCh <- true
		case floor := <-eventAtFloorCh:
			eventAtFloor(state, floor, shouldStopCh)
		case directionAndFloor := <-nextDirectionCh:
			state = eventNewDirection(id, state, directionAndFloor, startDoorTimerCh)
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
		fallthrough
	case STATE_STUCK:
		fmt.Println("Shouldstop satt\n")
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

func eventNewDirection(id string, state State, directionAndFloor []int, startDoorTimerCh chan<- bool) State {
	switch state {
	case STATE_STUCK:
		return STATE_MOVING
	case STATE_MOVING:
		fallthrough
	case STATE_IDLE:
		if directionAndFloor[0] == 0 {
			fmt.Println("Nå har vi stoppet")
			startDoorTimerCh <- true
			return STATE_DOOR_OPEN
		} else {
			fmt.Println("STATE_MOVING\n")
			return STATE_MOVING
		}
	}
	//fmt.Printf("eventND returnerer: %d\n", state)
	return state
}
