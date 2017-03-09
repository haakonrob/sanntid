package main

import (
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	//"./network/ring"
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
const subnet = "localhost"
//const subnet = "sanntidsal"

func main() {
	// The id is either 4th number of the local IPv4, or the PID of the 
	// process, depending on the specified subnet.
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

	/* network monitor */
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	bcastMsg := local.ID+"-"+local.IP
	go peers.Transmitter(15647, bcastMsg, subnet, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	/* ring network */
	ringUpdateCh := make(chan [2]string)
	// go ring.Node(incomingCh, outgoingCh, ringUpdateCh,
	/*helloTx := make(chan string)
	helloRx := make(chan string)
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)
	// The example message. We just send one of these every second.
	go func() {
		helloMsg := "Hello from..." + local.ID
		for {
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()
	*/
	
	fmt.Println("Started")
	var activePeers = make([]Peer, 0, MAX_NUM_PEERS)
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
			
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			
		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		}
	}
}
