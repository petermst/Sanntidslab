package Def

import (
	"sync"
)

type State int

const (
	STATE_IDLE State = 0
	STATE_MOVING
	STATE_DOOR_OPEN
	STATE_STUCK
)

const (
	N_BUTTONS int = 3
	N_FLOORS      = 4
)

const (
	MOTOR_DOWN int = -1
	MOTOR_IDLE
	MOTOR_UP
)

const (
	DOOR_CLOSE int = 0
	DOOR_OPEN      = 1
)

type driverState struct {
	id        string
	lastFloor int
	direction int
}

type NewOrLostPeer struct {
	id    string
	isNew bool
}

type ButtonIndicator struct {
	floor  int
	button int
	value  int
}

type QueueOperation struct {
	isAddOrder bool
	elevatorId string
	floor      int
	button     int
}

type Order struct {
	floor  int
	button int
}

type QueueMap struct {
	mux   sync.Mutex
	queue map[string][][]bool
}

type DriverStatesMap struct {
	mux    sync.Mutex
	states map[string][]int
}

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}
