package FSM

import (
	. "./Driver"
	"fmt"
	"os"
	"time"
)

type State int

const (
	STATE_IDLE State = 0
	STATE_MOVING
	STATE_DOOR_OPEN
	STATE_STUCK
)

func RunFSM(setMotorDirection chan int, startDoorTimer chan bool, eventElevatorStuck chan<- bool, eventAtFloor <-chan int, eventDoorTimeout <-chan bool, nextDirection <-chan int, shouldStop chan<- int, getNextDirection chan<- bool, elevatorStuckUpdateQueue chan<- bool) {
	state := STATE_IDLE

	for {
		select {
		case <-eventElevatorStuck:
			state = STATE_STUCK
			//elevatorStuckUpdateQueue <- true
		case floor := <-eventAtFloor:
			eventAtFloor(state, floor, shouldStop)
		case direction <- nextDirection:
			state = eventNewDirection(state, direction, startDoorTimer, setMotorDirection)
			setMotorDirection <- direction
		case <-eventDoorTimeout:
			state = eventDoorTimeout()
			getNextDirection <- true
		}
	}
}

func eventAtFloor(state State, floor int, shouldStop chan<- int) {
	switch state {
	case STATE_MOVING:
		shouldStop <- floor
	default:
		//FEIL
	}

}

func eventDoorTimeout() State {
	switch state {
	case STATE_DOOR_OPEN:

		return STATE_IDLE
	default:
		//FEIL
	}
}

func eventNewDirection(state State, direction int, startDoorTimer chan<- bool, setMotorDirection chan<- int) State {
	switch state {
	case STATE_MOVING:
	case STATE_IDLE:
		if direction == 0 {
			setMotorDirection <- direction
			startDoorTimer <- true
			return STATE_DOOR_OPEN
		} else {
			setMotorDirection <- direction
			return STATE_MOVING
		}
	default:
		//FEIL
	}

}
