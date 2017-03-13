type ChannelsQueueNetwork struct{
	MessageSentCh chan QueueOperation
	UpdateQueueSizeCh chan NewOrLostPeer
	IncommingQueueUpdateCh chan QueueOperation
	OutgoingQueueUpdateCh chan QueueOperation
	IncommingDriverStateUpdateCh chan DriverState 
	OutgoingDriverStateUpdateCh chan DriverState
}

type ChannelsQueueFSM struct{
	NextDirectionCh chan []int
	ShoudStopCh chan int
	GetNextDirectionCh chan bool
	ElevatorStuckUpdateQueueCh chan bool
}

type ChannelsQueueDriver struct{
	SetButtonIndicatorCh chan ButtonIndicator 
	CalcOptimalElevatorCh chan Order
	UpdateQueueCh chan QueueOperation
	IsDoorOpenCh chan bool
	IsDoorOpenResponseCh chan bool
}

type ChannelsFSMDriver struct{
	EventElevatorStuckCh chan bool
	EventAtFloorCh chan int
	EventDoorTimeoutCh chan bool
	SetMotorDirectionCh chan int
	StartDoorTimerCh chan bool
}

//Declaring new channels in main:

var channelsQueueNetwork ChannelsQueueNetwork 
channelsQueueNetwork.MessageSentCh := make(chan QueueOperation, 1)
channelsQueueNetwork.UpdateQueueSizeCh := make(chan NewOrLostPeer, 1)
channelsQueueNetwork.IncommingQueueUpdateCh := make(chan QueueOperation, 1)
channelsQueueNetwork.OutgoingQueueUpdateCh := make(chan QueueOperation, 1)
channelsQueueNetwork.IncommingDriverStateUpdateCh := make(chan DriverState, 1)
channelsQueueNetwork.OutgoingDriverStateUpdateCh := make(chan DriverState, 1)

var channelsQueueFSM ChannelsQueueFSM
channelsQueueFSM.NextDirectionCh := make(chan []int, 1) 
channelsQueueFSM.ShoudStopCh := make(chan int, 1) 
channelsQueueFSM.GetNextDirectionCh := make(chan bool, 1)
channelsQueueFSM.ElevatorStuckUpdateQueueCh := make(chan bool, 1) 

var channelsQueueDriver ChannelsQueueDriver
channelsQueueDriver.SetButtonIndicatorCh := make(chan ButtonIndicator, 1)
channelsQueueDriver.CalcOptimalElevatorCh := make(chan Order, 1)
channelsQueueDriver.UpdateQueueCh := make(chan QueueOperation, 1) 
channelsQueueDriver.IsDoorOpenCh := make(chan bool, 1)
channelsQueueDriver.IsDoorOpenResponseCh := make(chan bool, 1)

var channelsFSMDriver ChannelsFSMDriver
channelsFSMDriver.EventElevatorStuckCh := make(chan bool, 1)
channelsFSMDriver.EventAtFloorCh := make(chan int, 1) 
channelsFSMDriver.EventDoorTimeoutCh := make(chan bool, 1)
channelsFSMDriver.SetMotorDirectionCh := make(chan int, 1)   
channelsFSMDriver.StartDoorTimerCh := make(chan bool, 1)

