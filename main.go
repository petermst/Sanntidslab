package main

import (
	. "./Def"
	. "./Driver"
	. "./FSM"
	. "./Network"
	. "./Queue"
	"fmt"
)

func main() {

	initFloor := InitializeElevator()

	id := InitializeNetwork()
	fmt.Printf("My ID is: %s\n", id)

	var channelsQueueNetwork ChannelsQueueNetwork
	channelsQueueNetwork.MessageSentCh = make(chan QueueOperation, 1)
	channelsQueueNetwork.UpdateQueueSizeCh = make(chan NewOrLostPeer, 1)
	channelsQueueNetwork.IncomingQueueUpdateCh = make(chan QueueOperation, 1)
	channelsQueueNetwork.OutgoingQueueUpdateCh = make(chan QueueOperation, 1)
	channelsQueueNetwork.IncomingDriverStateUpdateCh = make(chan DriverState, 1)
	channelsQueueNetwork.OutgoingDriverStateUpdateCh = make(chan DriverState, 1)

	var channelsQueueFSM ChannelsQueueFSM
	channelsQueueFSM.NextDirectionCh = make(chan []int, 1)
	channelsQueueFSM.ShouldStopCh = make(chan int, 1)
	channelsQueueFSM.GetNextDirectionCh = make(chan bool, 1)
	channelsQueueFSM.ElevatorStuckRedistributeQueueCh = make(chan bool, 1)

	var channelsQueueDriver ChannelsQueueDriver
	channelsQueueDriver.SetButtonIndicatorCh = make(chan ButtonIndicator, 1)
	channelsQueueDriver.CalcOptimalElevatorCh = make(chan Order, 1)
	channelsQueueDriver.UpdateQueueCh = make(chan QueueOperation, 1)
	channelsQueueDriver.IsDoorOpenCh = make(chan bool, 1)
	channelsQueueDriver.IsDoorOpenResponseCh = make(chan bool, 1)

	var channelsFSMDriver ChannelsFSMDriver
	channelsFSMDriver.EventElevatorStuckCh = make(chan bool, 1)
	channelsFSMDriver.EventAtFloorCh = make(chan int, 1)
	channelsFSMDriver.EventDoorTimeoutCh = make(chan bool, 1)
	channelsFSMDriver.SetMotorDirectionCh = make(chan int, 1)
	channelsFSMDriver.StartDoorTimerCh = make(chan bool, 1)

	go RunDriver(id, channelsQueueDriver, channelsFSMDriver)
	go RunFSM(id, channelsQueueFSM, channelsFSMDriver)
	go RunNetwork(id, channelsQueueNetwork)
	go RunQueue(id, initFloor, channelsQueueNetwork, channelsQueueFSM, channelsQueueDriver)

	select {}
}
