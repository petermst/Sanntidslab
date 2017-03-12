package Def

import (
	"sync"
)

type State int

const (
	STATE_IDLE      State = 0
	STATE_MOVING    State = 1
	STATE_DOOR_OPEN State = 2
	STATE_STUCK     State = 3
)

const (
	N_BUTTONS int = 3
	N_FLOORS      = 4
)

const (
	MOTOR_DOWN int = -1
	MOTOR_IDLE int = 0
	MOTOR_UP   int = 1
)

const (
	DOOR_CLOSE int = 0
	DOOR_OPEN      = 1
)

const (
	IS_STOPPED int = 1
	IS_MOVING = 0
)

type DriverState struct {
	Id        string
	LastFloor int
	Direction int
	IsStopped int
}

type NewOrLostPeer struct {
	Id    string
	IsNew bool
}

type ButtonIndicator struct {
	Floor  int
	Button int
	Value  int
}

type QueueOperation struct {
	IsAddOrder bool
	ElevatorId string
	Floor      int
	Button     int
}

type Order struct {
	Floor  int
	Button int
}

type QueueMap struct {
	Mux   sync.Mutex
	Queue map[string][][]bool
}

type DriverStatesMap struct {
	Mux    sync.Mutex
	States map[string][]int
}

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

type Redistribute struct{
	Id string
	ShouldDelete bool
}