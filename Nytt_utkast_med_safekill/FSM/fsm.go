package FSM

import (
	. "../Def"
)

func RunFSM(id string, chQF ChannelsQueueFSM, chFD ChannelsFSMDriver) {

	state := STATE_IDLE

	for {

		select {
		case <-chFD.EventElevatorStuckCh:
			state = STATE_STUCK
			chQF.ElevatorStuckRedistributeQueueCh <- true

		case floor := <-chFD.EventAtFloorCh:
			eventAtFloor(state, floor, chQF.ShouldStopCh, chQF.GetNextDirectionCh)

		case directionAndFloor := <-chQF.NextDirectionCh:
			state = eventNewDirection(id, state, directionAndFloor, chFD.StartDoorTimerCh)
			chFD.SetMotorDirectionCh <- directionAndFloor[0]

		case <-chFD.EventDoorTimeoutCh:
			state = eventDoorTimeout(state)
			chQF.GetNextDirectionCh <- true
		}
	}
}

func eventAtFloor(state State, floor int, shouldStopCh chan<- int, getNextDirectionCh chan<- bool) State {
	switch state {
	case STATE_STUCK:
		fallthrough
	case STATE_MOVING:
		shouldStopCh <- floor
		return state

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

func eventNewDirection(id string, state State, directionAndFloor []int, startDoorTimerCh chan<- bool) State {
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
