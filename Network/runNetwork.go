package main

import (
	"./bcast"
	"./localip"
	"./peers"
	"fmt"
	"os"
	"time"
)

type isAliveMessage struct {
	lastFloor int
	direction int
}

type updateQueueMessage struct {
	operation bool
	elevator  int
	floor     int
	button    int
}

func runNetwork(elevatorID string, incomingMSG chan<- message, outgoingMSG <-chan message) {

	//make channel for receiving peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)

	//make channel for enableling transmitter for peer update
	peerTxEnable := make(chan bool)

	//goroutines for receiving and transmitting peerupdates
	go peers.Transmitter(10808, elevatorID, peerTxEnable)
	go peers.Receiver(10808, peerUpdateCh)

	//channels for sending and receiving custom data types, message for queueupdate
	IsAliveT := make(chan isAliveMessage)
	IsaliveR := make(chan isAliveMessage)

	updateT := make(chan updateQueueMessage)
	updateR := make(chan updateQueueMessage)

	//goroutines for transmitting and receiving custom data types
	go bcast.Transmitter(30008, IsAliveT, updateT)
	go bcast.Receiver(30008, IsaliveR, updateR)

	//function that iteratively sends a message to tell other elevators that it lives
	go func() {
		//Update elevators on the network
		updateMSG := isAliveMessage{, 0}
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
		case newMSG := <-Rx:
			receiveMessage(newMSG, incomingMSG)
		case Tx := <-outgoingMSG:

		}
	}

}

func receiveMessage(Message message, incomingMSG <-chan message) {

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
