package Def

import (
	"sync"
)

type State int

type DriverState struct {
	ElevatorId string
	LastFloor  int
	Direction  int
	IsStopped  int
}

type NewOrLostPeer struct {
	ElevatorId string
	IsNew      bool
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
	Queue map[string][][]bool //Matrix with rows = floor, columns = button
}

type DriverStatesMap struct {
	Mux    sync.Mutex
	States map[string][]int //Index 0: Floor, index 1: Direction, index 2: IsStopped
}

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}
