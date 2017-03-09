package FSM

import (
	. "./Driver"
	. "./Queue"
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

func RunFSM(id string, setMotorDirection chan int, startDoorTimer chan bool, eventElevatorStuck chan<- bool, eventAtFloor <-chan int, eventDoorTimeout <-chan bool, nextDirection <-chan []int, shouldStop chan<- int, getNextDirection chan<- bool, elevatorStuckUpdateQueue chan<- bool, updateQueue chan<- QueueOperation) {
	state := STATE_IDLE

	for {
		select {
		case <-eventElevatorStuck:
			state = STATE_STUCK
			elevatorStuckUpdateQueue <- true
		case floor := <-eventAtFloor:
			eventAtFloor(state, floor, shouldStop)
		case directionAndFloor <- nextDirection:
			state = eventNewDirection(id, state, directionAndFloor, startDoorTimer, setMotorDirection)
			setMotorDirection <- directionAndFloor[0]
		case <-eventDoorTimeout:
			state = eventDoorTimeout(state)
			getNextDirection <- true
		}
	}
}

func eventAtFloor(state State, floor int, shouldStop chan<- int) {
	switch state {
	case STATE_MOVING:
		shouldStop <- floor
	default:
		return state
	}

}

func eventDoorTimeout(state State) State {
	switch state {
	case STATE_DOOR_OPEN:

		return STATE_IDLE
	default:
		return state
	}
}

func eventNewDirection(id string, state State, directionAndFloor []int, startDoorTimer chan<- bool, setMotorDirection chan<- int) State {
	switch state {
	case STATE_MOVING:
	case STATE_IDLE:
		if directionAndFloor[0] == 0 {
			startDoorTimer <- true
			QueOpe := QueueOperation{false, id, directionAndFloor[1], 0}
			updateQueue <- QueOpe
			return STATE_DOOR_OPEN
		} else {
			return STATE_MOVING
		}
	default:
		return state
	}

}
