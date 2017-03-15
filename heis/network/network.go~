package network

import (
	"./bcast"
	"./localip"
	"./peers"
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
const subnet = "sanntidsal" //or localhost
const loopBack = true
const peerPort = 20005
const ringport = 20006
const UDPPasscode = "svekonrules"

func Monitor(statusCh chan string, incomingCh chan []byte, outgoingCh chan []byte) {

	/*
		The id is either 4th number of the local IPv4, or the PID of the
		process, depending on the specified subnet and loopback mode.
	*/
	var local Peer

	if loopBack {
		local.ID = fmt.Sprintf("%d", os.Getpid())
	}
	IP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		IP = "DISCONNECTED"
	}
	local.IP = IP

	/* Start monitoring network over UDP */
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	bcastMsg := UDPPasscode + "_" + local.ID + "-" + local.IP
	go peers.Transmitter(peerPort, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(peerPort, UDPPasscode, peerUpdateCh)

	/* Ring network */

	targetCh := make(chan string)
	go bcast.Transmitter(ringport, targetCh, outgoingCh)
	go bcast.Receiver(ringport, local.ID, incomingCh)

	fmt.Println("Network module started up PID", local.ID)
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
				if len(activePeers) > 1 {
					online = true
					i := getLocalPeerIndex(local, activePeers)
					next_i := (i + 1) % len(activePeers)
					targetCh <- activePeers[next_i].ID

				} else {
					online = false
				}
				update = false
				msg := fmt.Sprintf("%t_%s_", online, local.ID)
				for _, pr := range activePeers {
					msg = msg + pr.ID + "-"
				}
				statusCh <- msg[0:(len(msg) - 1)]
			}
		}

	}
}

func getLocalPeerIndex(ID Peer, list []Peer) int {
	i := 0
	for i < len(list) {
		if ID == list[i] {
			break
		}
		i++
	}
	return i
}
