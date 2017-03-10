package network

import (
	"./localip"
	"./peers"
	"./ring"
	"fmt"
	"os"
	"time"
	"strings"
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
const loopBack = false
const peerPort = 20024
const TCPPort = 20000
const UDPPasscode = "svekonrules"

func Monitor(coordinatorCh chan string, incomingCh chan string, outgoingCh chan string) {

	/* 
	The id is either 4th number of the local IPv4, or the PID of the 
	process, depending on the specified subnet and loopback mode.
	*/
	var local Peer
	
	ringPort := TCPPort
	if loopBack {
		local.ID = fmt.Sprintf("%d",os.Getpid())
		ringPort = os.Getpid()
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
	bcastMsg := UDPPasscode + "_" + local.ID+"-"+local.IP
	go peers.Transmitter(peerPort, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(peerPort, UDPPasscode, peerUpdateCh)

	/* Ring network */
	ringNextCh := make(chan string)
	ringPrevCh := make(chan string)
	go ring.NextNode(outgoingCh, ringNextCh)
	go ring.PrevNode(incomingCh, ringPrevCh, ringPort)

	fmt.Println("Network module started up PID", local.ID)
	// every node will send a reply when it has been successfully updated. OK or ERROR.
	
	var localIndex int
	var activePeers = make([]Peer, 0, MAX_NUM_PEERS)
	update := false

	for {
		select {

		case p := <-peerUpdateCh:
			activePeers = make([]Peer, len(p.Peers), MAX_NUM_PEERS)
			for i, pr := range p.Peers {
				newData := strings.Split(pr,"-")
				activePeers[i] = Peer{newData[0], newData[1]}
			}
			update = true
			time.Sleep(time.Millisecond*200)
			
		case nn := <-ringNextCh:
			if nn == "ERROR"{
				//fmt.Println("NextNode conn closed")
				update = true
			}

		default:
			if update && len(activePeers) > 1 {
				//fmt.Println(activePeers)
				localIndex = getLocalPeerIndex(local, activePeers)
				next_i := (localIndex + 1) % len(activePeers)
				ringPrevCh<- "RESET"
				/****TEMPORARY*****/
				ringNextCh<- fmt.Sprintf("%s:%d", activePeers[next_i].IP, ringPort)
				//ringNextCh<- fmt.Sprintf("%s:%i", activePeers[next_i].IP, port)
				/******************/
				if "OK" != <-outgoingCh {
					fmt.Println("NextNode closed")
				}
				coordinatorCh<- fmt.Sprintf("%d", localIndex)
				update = false
				//get next node addr, send to nextnode, reset prevNode
			}
		}
		
		
	}
}

func getLocalPeerIndex(ID Peer, list []Peer)(int){
	i:=0
	for i<len(list) {
		if ID == list[i]{
			break;
		}
		i++
	}
	return i
}

