package Network

import (
	"fmt"
	"os"
	"time"
)

type driverState struct {
	id        string
	lastFloor int
	direction int
}

type updateQueueMessage struct {
	operation bool
	elevator  int
	floor     int
	button    int
}

func runNetwork(elevatorID string, updatePeersOnQueue chan<- driverState, incomingMSG chan<- message, outgoingMSG <-chan message, transmitUpdate <-chan driverState) {

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

	updateTransmit := make(chan updateQueueMessage)
	updateReceive := make(chan updateQueueMessage)

	//goroutines for transmitting and receiving custom data types
	go TransmitterBcast(30008, updateTransmit)
	go ReceiverBcast(30008, updateReceive)

	//function that iteratively sends a message to tell other elevators that it lives
	go func() {
		//Update elevators on the network
		updateMSG := driverState{, 0}
		updateMSG.id = elevatorID
		for {
			Tx <- updateMSG
			time.Sleep(time.Second)
		}

	}()

	for {
		select {
		case peerUpdate := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
		case newIsAlive := <- IsaliveR:
			receiveMessage(newIsAlive, incomingMSG)
		case Tx := <-outgoingMSG:

		}
	}

}

func receiveIsAlive(Message driverState, incomingMSG <-chan message) {
	
}

func sendMessage(Message) {

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
