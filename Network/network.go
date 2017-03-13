package Network

import (
	. "../Def"
	"fmt"
	//"os"
)

/*
Credit to Anders Rønning Petersen
From https://github.com/TTK4145/Network-go
*/

func RunNetwork(id string, chQN ChannelsQueueNetwork) {

	//make channel for receiving peer updates
	peerUpdateCh := make(chan PeerUpdate, 1)

	//make channel for enableling transmitter for peer update
	peerTxEnableCh := make(chan bool, 1)

	//goroutines for receiving and transmitting peerupdates
	go TransmitterPeers(10808, id, peerTxEnableCh)
	go ReceiverPeers(10808, peerUpdateCh)

	//channels for sending and receiving custom data types, message for queue update and driverstate update
	transmitDriverStateUpdateCh := make(chan DriverState, 1)
	receiveDriverStateUpdateCh := make(chan DriverState, 1)

	transmitQueueUpdateCh := make(chan QueueOperation, 1)
	copyTransmitQueueUpdateCh := make(chan QueueOperation, 1)
	receiveQueueUpdateCh := make(chan QueueOperation, 1)

	//goroutines for transmitting and receiving custom data types
	go TransmitterBcast(31345, chQN.MessageSentCh, copyTransmitQueueUpdateCh, transmitQueueUpdateCh, transmitDriverStateUpdateCh)
	go ReceiverBcast(31345, receiveQueueUpdateCh, receiveDriverStateUpdateCh)

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			updateNumberOfPeers(id, peerUpdate, chQN.UpdateQueueSizeCh)

		case driverstateOutgoing := <-chQN.OutgoingDriverStateUpdateCh:
			transmitDriverStateUpdateCh <- driverstateOutgoing

		case driverStateIncoming := <-receiveDriverStateUpdateCh:
			chQN.IncomingDriverStateUpdateCh <- driverStateIncoming

		case queueUpdateOutgoing := <-chQN.OutgoingQueueUpdateCh:
			transmitQueueUpdateCh <- queueUpdateOutgoing
			copyTransmitQueueUpdateCh <- queueUpdateOutgoing

		case queueUpdateIncoming := <-receiveQueueUpdateCh:
			chQN.IncomingQueueUpdateCh <- queueUpdateIncoming
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
		id = fmt.Sprintf("peer-%s", localIP)
	}
	return id
}

func updateNumberOfPeers(id string, peerUpdate PeerUpdate, updateQueueSizeCh chan<- NewOrLostPeer) {
	var p NewOrLostPeer
	for i := range peerUpdate.Lost {
		p.ElevatorId = peerUpdate.Lost[i]
		p.IsNew = false
		updateQueueSizeCh <- p
	}

	if peerUpdate.New != "" {
		p.ElevatorId = peerUpdate.New
		p.IsNew = true
		if peerUpdate.New != id {
			updateQueueSizeCh <- p
		}
	}

	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
	fmt.Printf("  New:      %q\n", peerUpdate.New)
	fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
}
