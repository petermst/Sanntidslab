package Network

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"
	. "../Def"
)



const interval = 10 * time.Millisecond
const timeout = 50 * time.Millisecond

func TransmitterPeers(port int, id string, transmitEnableCh <-chan bool, newPeerTransmitMessageCh <-chan DriverState) {
	
	currentDriverState := DriverState{id,1,-1}
	peerTransmitMessageCh := make(chan DriverState)

	conn := DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	selectCases := make([]reflect.SelectCase, 1)
	typeNames := make([]string, 1)

	selectCases[0] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(peerTransmitMessageCh)}
	typeNames[0] = reflect.TypeOf(peerTransmitMessageCh).Elem().String()

	enable := true

	for {
		select {
		case enable = <-transmitEnableCh:
		case currentDriverState = <- newPeerTransmitMessageCh:
		case <-time.After(interval):
		}

		peerTransmitMessageCh <- currentDriverState

		if enable {
			chosen, value, _ := reflect.Select(selectCases)
			buf, _ := json.Marshal(value.Interface())
			conn.WriteTo([]byte(typeNames[chosen]+string(buf)), addr)
		}
	}
}

func ReceiverPeers(port int, peerUpdateCh chan<- PeerUpdate, updatePeersOnQueueCh chan<- DriverState) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := DialBroadcastUDP(port)

	newMessageCh := make(chan DriverState)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])
		T := reflect.TypeOf(newMessageCh).Elem()
		typename := T.String()
		if strings.HasPrefix(string(buf[:n])+"{", typename) {
			v := reflect.New(T)
			json.Unmarshal(buf[len(typename):n], v.Interface())

			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(newMessageCh),
				Send: reflect.Indirect(v),
			}})
		}

		message := <- newMessageCh

		// Adding new connection
		p.New = ""
		if message.Id != "" {
			if _, idExists := lastSeen[message.Id]; !idExists {
				p.New = message.Id
				updated = true
			}
			lastSeen[message.Id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
		updatePeersOnQueueCh <- message
	}
}
