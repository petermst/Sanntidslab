package Network

import (
	. "../Def"
	"fmt"
	"os"
	//"time"
)

func runNetwork(elevatorID string, updatePeersOnQueueCh chan<- DriverState, updateQueueSizeCh chan<- NewOrLostPeer, incomingMessageCh chan<- QueueOperation, outgoingMessageCh <-chan QueueOperation, peersTransmitMessageCh chan DriverState, messageSentCh chan<- QueueOperation) {

	//make channel for receiving peer updates
	peerUpdateCh := make(chan PeerUpdate)

	//make channel for enableling transmitter for peer update
	peerTxEnableCh := make(chan bool)

	//goroutines for receiving and transmitting peerupdates
	go TransmitterPeers(10808, elevatorID, peerTxEnableCh, peersTransmitMessageCh)
	go ReceiverPeers(10808, peerUpdateCh, updatePeersOnQueueCh)

	//channels for sending and receiving custom data types, message for queueupdate

	broadcastTransmitMessageCh := make(chan QueueOperation)
	broadcastReceiveMessageCh := make(chan QueueOperation)

	//goroutines for transmitting and receiving custom data types
	go TransmitterBcast(30008, messageSentCh, broadcastTransmitMessageCh)
	go ReceiverBcast(30008, broadcastReceiveMessageCh)

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			updateNumberOfPeers(peerUpdate, updateQueueSizeCh)
		case messageGoingOut:= <- outgoingMessageCh:
			broadcastTransmitMessageCh <- messageGoingOut
		case messageGoingIn:= <- broadcastReceiveMessageCh:
			incomingMessageCh <-messageGoingIn
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
	if peerUpdate.New != ""{
		p.Id = peerUpdate.New
		p.IsNew = true 
		updateQueueSizeCh <- p
	}
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
	fmt.Printf("  New:      %q\n", peerUpdate.New)
	fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
}
