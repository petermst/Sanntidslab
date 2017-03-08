package Network

import (
	. "./Queue"
	"fmt"
	"os"
	"time"
)

type driverState struct {
	id        string
	lastFloor int
	direction int
}

type NewOrLostPeer struct{
	id string
	isNew bool
}

func runNetwork(elevatorID string, updatePeersOnQueue chan<- driverState, updateQueueSize chan<- NewOrLostPeer, incomingMSG chan<- QueueOperation, outgoingMSG <-chan QueueOperation, peersTransmitMSG <-chan driverState, messageSent chan<- QueueOperation) {

	var lastFloor = -1
	var direction = -1

	//make channel for receiving peer updates
	peerUpdateCh := make(chan PeerUpdate)

	//make channel for enableling transmitter for peer update
	peerTxEnable := make(chan bool)

	//goroutines for receiving and transmitting peerupdates
	go TransmitterPeers(10808, elevatorID, peerTxEnable, transmitUpdate)
	go ReceiverPeers(10808, peerUpdateCh, updatePeersOnQueue)

	//channels for sending and receiving custom data types, message for queueupdate

	broadcastTransmitMSG := make(chan QueueOperation)
	broadcastReceiveMSG := make(chan QueueOperation)

	//goroutines for transmitting and receiving custom data types
	go TransmitterBcast(30008, messageSent, broadcastTransmitMSG)
	go ReceiverBcast(30008, broadcastReceiveMSG)

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			updateNumberOfPeers(peerUpdate, updateQueueSize)
		case newIsAlive := <- IsaliveR:
			receiveMessage(newIsAlive, incomingMSG)
		case broadcastTransmitMSG := <-outgoingMSG:

		case incomingMSG <- broadcastReceiveMSG
		}
	}

}

func receiveIsAlive(Message driverState, incomingMSG <-chan message) {
	
}

func InitializeNetwork() string {
	var id string
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	return id
}


func updateNumberOfPeers(peerUpdate PeerUpdate, updateQueueSize chan<- NewOrLostPeer) {
	p NewOrLostPeer
	for i := range peerUpdate.Lost {
		p.id = peerUpdate[i]
		p.isNew = false
		updateQueueSize <- p
	}
	if peerUpdate.New != ""{
		p.id = peerUpdate.New
		p.isNew = true 
		updateQueueSize <- p
	}
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
	fmt.Printf("  New:      %q\n", peerUpdate.New)
	fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
}
