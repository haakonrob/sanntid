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
//const subnet = "localhost"
const subnet = "sanntidsal"
const UDPPasscode = "svekonrules"

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
		local.ID = fmt.Sprintf("%d",os.Getpid())
		//local.ID = strings.Split(IP,".")[3]
	default:
		local.IP = IP
		local.ID = fmt.Sprintf("%d",os.Getpid())
	}

	/* Start monitoring network over UDP */
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	bcastMsg := UDPPasscode + "_" + local.ID+"-"+local.IP
	go peers.Transmitter(20024, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(20024, UDPPasscode, peerUpdateCh)
	fmt.Println("Started network monitor")

	/* Ring network */
	ringNextCh := make(chan string)
	ringPrevCh := make(chan string)
	go ring.NextNode(outgoingCh, ringNextCh)
	go ring.PrevNode(incomingCh, ringPrevCh, local.ID)
	fmt.Println("Started TCP ring")
	// every node will send a reply when it has been successfully updated. OK or ERROR.
	
	//var localIndex int
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
		}
		
		if update && len(activePeers) > 1 {
			localIndex := getLocalPeerIndex(local, activePeers)
			next_i := (localIndex + 1) % len(activePeers)
			ringPrevCh<- "RESET"
			/****TEMPORARY*****/
			//str := fmt.Sprintf("%s:%s", activePeers[next_i].IP, activePeers[next_i].ID)
			//fmt.Println("Updated TCP addr", str)
			ringNextCh<- fmt.Sprintf("%s:%s", activePeers[next_i].IP, activePeers[next_i].ID)
			//ringNextCh<- fmt.Sprintf("%s:%i", activePeers[next_i].IP, port)
			/******************/
			//fmt.Println(<-ringNextCh)
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

