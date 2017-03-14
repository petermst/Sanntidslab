package Def

type ChannelsQueueNetwork struct {
	MessageSentCh               chan QueueOperation
	UpdateQueueSizeCh           chan NewOrLostPeer
	IncomingQueueUpdateCh       chan QueueOperation
	OutgoingQueueUpdateCh       chan QueueOperation
	IncomingDriverStateUpdateCh chan DriverState
	OutgoingDriverStateUpdateCh chan DriverState
}

type ChannelsQueueFSM struct {
	NextDirectionCh                  chan []int
	ShouldStopCh                     chan int
	GetNextDirectionCh               chan bool
	ElevatorStuckRedistributeQueueCh chan bool
}

type ChannelsQueueDriver struct {
	SetButtonIndicatorCh  chan ButtonIndicator
	CalcOptimalElevatorCh chan Order
	UpdateQueueCh         chan QueueOperation
	IsDoorOpenCh          chan bool
	IsDoorOpenResponseCh  chan bool
}

type ChannelsFSMDriver struct {
	EventElevatorStuckCh chan bool
	EventAtFloorCh       chan int
	EventDoorTimeoutCh   chan bool
	SetMotorDirectionCh  chan int
	StartDoorTimerCh     chan bool
}
