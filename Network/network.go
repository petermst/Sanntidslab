package Network

import (
	. "../Def"
	"fmt"
	"os"
	//"time"
)

func RunNetwork(elevatorID string, updateQueueSizeCh chan<- NewOrLostPeer, messageSentCh chan<- QueueOperation, outgoingQueueUpdateCh <-chan QueueOperation, incomingQueueUpdateCh chan<- QueueOperation, outgoingDriverStateUpdateCh <-chan DriverState, incomingDriverStateUpdateCh chan<- DriverState) {

	//make channel for receiving peer updates
	peerUpdateCh := make(chan PeerUpdate, 1)

	//make channel for enableling transmitter for peer update
	peerTxEnableCh := make(chan bool, 1)

	//goroutines for receiving and transmitting peerupdates
	go TransmitterPeers(10808, elevatorID, peerTxEnableCh)
	go ReceiverPeers(10808, peerUpdateCh)

	//channels for sending and receiving custom data types, message for queue update and driverstate update

	transmitDriverStateUpdate := make(chan DriverState, 1)
	receiveDriverStateUpdate := make(chan DriverState, 1)

	transmitQueueUpdate := make(chan QueueOperation, 1)
	copyTransmitQueueUpdate := make(chan QueueOperation, 1)
	receiveQueueUpdate := make(chan QueueOperation, 1)

	//goroutines for transmitting and receiving custom data types
	go TransmitterBcast(32345, messageSentCh, copyTransmitQueueUpdate, transmitQueueUpdate, transmitDriverStateUpdate)
	go ReceiverBcast(32345, receiveQueueUpdate, receiveDriverStateUpdate)

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			updateNumberOfPeers(peerUpdate, updateQueueSizeCh)
		case driverstateOutgoing := <- outgoingDriverStateUpdateCh:
			transmitDriverStateUpdate <- driverstateOutgoing
		case driverStateIncoming := <- receiveDriverStateUpdate:
			incomingDriverStateUpdateCh <- driverStateIncoming
		case queueUpdateOutgoing := <- outgoingQueueUpdateCh:
			transmitQueueUpdate <- queueUpdateOutgoing
			copyTransmitQueueUpdate <- queueUpdateOutgoing
		case queueUpdateIncoming := <-receiveQueueUpdate:
			if queueUpdateIncoming.ElevatorId != elevatorID {
				fmt.Printf("Dette legges inn i incomingMSG: ID: %s , isAdd: %t , floor: %d, button %d\n", messageGoingIn.ElevatorId, messageGoingIn.IsAddOrder, messageGoingIn.Floor, messageGoingIn.Button)
				incomingQueueUpdateCh <- queueUpdateIncoming
			}
		}
	}
}

func InitializeNetwork() string {
	var id string
	if id == "" {
		localIP, err := LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	return id
}

func updateNumberOfPeers(peerUpdate PeerUpdate, updateQueueSizeCh chan<- NewOrLostPeer) {
	var p NewOrLostPeer
	for i := range peerUpdate.Lost {
		p.Id = peerUpdate.Lost[i]
		p.IsNew = false
		updateQueueSizeCh <- p
	}
	if peerUpdate.New != "" {
		p.Id = peerUpdate.New
		p.IsNew = true
		updateQueueSizeCh <- p
	}
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
	fmt.Printf("  New:      %q\n", peerUpdate.New)
	fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
}
