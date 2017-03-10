package network

import (
	"./localip"
	"./peers"
	"./ring"
	"fmt"
	"os"
	//"time"
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
const subnet = "localhost"
//const subnet = "sanntidsal"

func Monitor(incomingCh chan string, outgoingCh chan string) {

	/* 
	The id is either 4th number of the local IPv4, or the PID of the 
	process, depending on the specified subnet.
	*/
	var local Peer
	IP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		IP = "DISCONNECTED"
	}

	switch subnet {
	case "sanntidsal":
		local.IP = IP
		local.ID = strings.Split(IP,".")[3]
	default:
		local.IP = IP
		local.ID = fmt.Sprintf("%d",os.Getpid())
	}

	/* Start monitoring network over UDP */
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	bcastMsg := local.ID+"-"+local.IP
	go peers.Transmitter(15647, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	/* Ring network */
	ringNextCh := make(chan string)
	ringPrevCh := make(chan string)
	go ring.NextNode(outgoingCh, ringNextCh, 20024)
	go ring.PrevNode(incomingCh, ringPrevCh, 20024)
	// every node will send a reply when it has been successfully updated. OK or ERROR.
	
	fmt.Println("Started network monitor")
	var localIndex int
	var activePeers = make([]Peer, 0, MAX_NUM_PEERS)
	update := false
	//update_lost_peers := false
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update!\n")
			activePeers = make([]Peer, len(p.Peers), MAX_NUM_PEERS)
			for i, pr := range p.Peers {
				newData := strings.Split(pr,"-")
				activePeers[i] = Peer{newData[0], newData[1]}
			}
			fmt.Println("Active Peers: ", activePeers)
			update = true
			/*if p.New != "" {
				update_new_peers = true
			}
			if len(p.Lost) > 0 {
				update_lost_peers = true
			}*/
			/*
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			*/
		case a := <-ringPrevCh:
			fmt.Printf("Received: %#v\n", a)
			ringNextCh<- a
			if (<-ringNextCh) == "OK"{
				fmt.Println("Sent succesfully")
			} else {
			 fmt.Println("Failed to send to next node")
			}
		}
		
		if update && len(activePeers) > 1 {
			localIndex = getLocalPeerIndex(local, activePeers)
			next_i := (localIndex + 1) % len(activePeers)
			ringPrevCh<- "RESET"
			ringNextCh<- activePeers[next_i].IP
			//get next node addr, send to nextnode, reset prevNode
			
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

