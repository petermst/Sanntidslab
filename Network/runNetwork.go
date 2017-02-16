package main

import (
	"./bcast"
	"./localip"
	"./peers"
	"fmt"
	"os"
	"time"
)

type message struct {
	Content string
	id      string
}

func runNetwork(incomingMSG chan<- message, outgoingMSG <-chan message) {

	//make elevator id
	var id string
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	//make channel for receiving peer updates
	peerUpdateCh := make(chan peers.PeerUpdate)

	//make channel for enableling transmitter for peer update
	peerTxEnable := make(chan bool)

	//goroutines for receiving and transmitting peerupdates
	go peers.Transmitter(10808, id, peerTxEnable)
	go peers.Receiver(10808, peerUpdateCh)

	//channels for sending and receiving custom data types, message for queueupdate
	Tx := make(chan message)
	Rx := make(chan message)

	//goroutines for transmitting and receiving custom data types
	go bcast.Transmitter(30008, Tx)
	go bcast.Receiver(30008, Rx)

	//function that iteratively sends a message to tell other elevators that it lives
	go func() {
		//Update elevators on the network
		updateMSG := message{"Jeg er en heis", 0}
		updateMSG.id = id
		for {
			Tx <- updateMSG
			time.Sleep(time.Second)
		}

	}()

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case newMSG := <-Rx:
			receiveMessage(newMSG, incomingMSG)

		case sendMSG := <-outgoingMSG:
			sendMessage(sendMSG)

		case peerUpdate := <-peerUpdateCh:

		}
	}

}

func sendMessage(Message message) {
	Tx <- Message
}

func receiveMessage(Message message, incomingMSG chan<- message) {

}
