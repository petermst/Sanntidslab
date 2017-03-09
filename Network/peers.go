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

func TransmitterPeers(port int, id string, transmitEnable <-chan bool, newPeerTransmitMSG  <-chan driverState) {
	
	currentDriverState := driverState{id,1,-1}
	peerTransmitMSG := make(chan driverState)

	conn := DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	selectCases := make([]reflect.SelectCase, 1)
	typeNames := make([]string, 1)

	selectCases[0] = reflect.SelectCase{
		Dir: reflect.SelectRecv,
		Chan: reflect.ValueOf(peerTransmitMSG)
	}
	typenames[0] = reflect.TypeOf(peerTransmitMSG).Elem().String()

	enable := true

	for {
		select {
		case enable = <-transmitEnable:
		case currentDriverState = <- newPeerTransmitMSG:
		case <-time.After(interval):
		}

		peerTransmitMSG <- currentDriverState

		if enable {
			chosen, value, _ := reflect.Select(selectCases)
			buf, _ := json.Marshal(value.Interface())
			conn.WriteTo([]byte(typeNames[chosen]+string(buf)), addr)
		}
	}
}

func ReceiverPeers(port int, peerUpdateCh chan<- PeerUpdate, updatePeersOnQueue chan<- driverState) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := DialBroadcastUDP(port)

	newMessage := make(chan driverState)
	var message driverState

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])
		T := reflect.TypeOf(newMessage).Elem()
		typename := T.String()
		if strings.HasPrefix(buf[:n], typename) {
			v := reflect.New(T)
			json.Unmarshal(buf[len(typename):n], v.Interface())

			reflect.Select([]reflect.SelectCase{{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(newMessage),
				Send: reflect.Indirect(v),
			}})
		}

		message <- newMessage

		// Adding new connection
		p.New = ""
		if message.id != "" {
			if _, idExists := lastSeen[message.id]; !idExists {
				p.New = message.id
				updated = true
			}
			lastSeen[message.id] = time.Now()
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
		updatePeersOnQueue <- message
	}
}
