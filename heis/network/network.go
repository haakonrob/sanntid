package network

import (
	"./ring"
	"./localip"
	"./peers"
	"fmt"			
	"os"
	"strings"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

type Peer struct {
	ID string
	IP string
}

type Status struct {
	Online bool
	LocalID string
	ActiveIDs [] string
}

const 
(
 	MAX_NUM_PEERS = 10
 	UDPPasscode = "svekonrules"
 	peerPort = 20005
)

var (
	netStat Status
	ringport = 20006
)

func Monitor(statusCh chan Status, loopBack bool, subnet string, incomingCh chan interface{}, outgoingCh chan interface{}) {
	/*
		The local id is either 4th number of the local IPv4, or the PID of the
		process, depending on the specified subnet and loopback mode.
	*/
	local := getLocalInfo(loopBack, subnet)

	/* Start monitoring network using UDP */
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	bcastMsg := fmt.Sprintf("%s_%s-%s:%d",UDPPasscode,local.ID, local.IP, ringport)
	go peers.Transmitter(peerPort, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(peerPort, UDPPasscode, peerUpdateCh)

	/* Start ring network using TCP */
	targetCh := make(chan string)
	go ring.Transmitter(targetCh, outgoingCh)
	go ring.Receiver(ringport, incomingCh)

	fmt.Println("Network module started up PID", local.ID)
	
	for {
		p := <-peerUpdateCh

		var activePeers []Peer
		var activeIDs []string
		for _, pr := range p.Peers {
			newData := strings.Split(pr, "-")
			activePeers = append(activePeers, Peer{newData[0], newData[1]} )
			activeIDs = append(activeIDs, newData[0])
		}

		if len(activePeers) > 1 {
			i := getLocalPeerIndex(local, activePeers)
			next_i := (i + 1) % len(activePeers)
			targetCh <- fmt.Sprintf("%s:%d", activePeers[next_i].IP, ringport)
			statusCh <- Status{true, local.ID, activeIDs}
		} else {
			statusCh <- Status{false, local.ID, activeIDs}
		}

		
	}
}


func getLocalInfo(loopBack bool, subnet string)(Peer){
	var local Peer
	IP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		IP = ""
	}

	local.IP = IP

	if loopBack {
		local.ID = fmt.Sprintf("%d", os.Getpid())
		ringport = os.Getpid()
	} else {
		fmt.Println(strings.Split(IP, "."))
		local.ID = strings.Split(IP, ".")[3]
	}
	return local
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
