package network

import (
	"./localip"
	"./peers"
	"./ring"
	"fmt"
	"os"
	"strings"
	_ "time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

type Peer struct {
	ID string
	IP string
}

const MAX_NUM_PEERS = 10

//const subnet = "sanntidsal" //or localhost
const UDPPasscode = "svekonrulesss"
const peerPort = 20005

var ringport = 20006

func Monitor(statusCh chan string, loopBack bool, subnet string, incomingCh chan interface{}, outgoingCh chan interface{}) {

	/*
		The id is either 4th number of the local IPv4, or the PID of the
		process, depending on the specified subnet and loopback mode.
	*/
	var local Peer

	IP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		IP = "DISCONNECTED"
	}
	local.IP = IP

	if loopBack {
		local.ID = fmt.Sprintf("%d", os.Getpid())
		ringport = os.Getpid()
	} else {
		fmt.Println(strings.Split(IP, "."))
		local.ID = strings.Split(IP, ".")[3]
	}

	/* Start monitoring network over UDP */
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	bcastMsg := fmt.Sprintf("%s_%s-%s:%d", UDPPasscode, local.ID, local.IP, ringport)
	go peers.Transmitter(peerPort, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(peerPort, UDPPasscode, peerUpdateCh)

	/* Ring network */
	targetCh := make(chan string)
	go ring.Transmitter(targetCh, outgoingCh)
	go ring.Receiver(ringport, incomingCh)

	fmt.Println("Network module started up. PID:", local.ID, ", IP: ", local.IP)
	// every node will send a reply when it has been successfully updated. OK or ERROR.

	var activePeers = make([]Peer, 0, MAX_NUM_PEERS)
	online := false
	update := false

	for {
		select {

		case p := <-peerUpdateCh:

			activePeers = make([]Peer, len(p.Peers), MAX_NUM_PEERS)
			for i, pr := range p.Peers {
				newData := strings.Split(pr, "-")
				activePeers[i] = Peer{newData[0], newData[1]}
			}
			update = true

		default:
			if update {
				update = false
				if len(activePeers) > 1 {
					online = true
					i := getLocalPeerIndex(local, activePeers)
					next_i := (i + 1) % len(activePeers)
					nextTarget := activePeers[next_i].IP
					targetCh <- nextTarget
				} else {
					online = false
				}
				msg := fmt.Sprintf("%t_%s_", online, local.ID)
				fmt.Printf(msg)
				for _, pr := range activePeers {
					msg = msg + pr.ID + "-"
				}
				statusCh <- msg[0 : (len(msg))-1]
			}
		}

	}
}

func getLocalPeerIndex(p Peer, list []Peer) int {
	i := 0
	for i < len(list) {
		if p.ID == list[i].ID {
			return i
		}
		i++
	}
	fmt.Println("My ID isn't in the list.")
	return i
}
