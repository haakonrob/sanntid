package networkmonitorTest
import (
	"net"
	"bufio"
	"time"
	"../localnet"
	"../ringnode"
	"fmt"
)

const (
	TCPPortIn = ":20024"
	TCPPortOut = ":20025"
	UDPPort = ":20023"
	UDPPasscode = "svekonrules"
	UDPTimeout = 100*time.Millisecond
)

func NetworkMonitor(packetChannel chan string, monitorChannel chan string){
	localnet.Init()
	//localIP, _ := localnet.IP()
	//broadcastIP, _ := localnet.BroadcastIP()
	//bcastMsg := UDPPasscode+"\n"+localIP+"\n"
	updateChannel := make(chan string)
	UDPChan := make(chan string)
	//UDPBroadcastDone := make(chan bool)
	//go UDPReceiver(UDPChan, UDPPasscode, UDPPort)
	//go UDPBroadcaster(UDPBroadcastDone, bcastMsg, broadcastIP, UDPPort)
	go ringnode.RingNode(packetChannel, updateChannel, TCPPortIn, TCPPortOut)
	//localnet.PeerUpdate("129.241.187.48")
	localnet.PeerUpdate("10.24.39.211"+TCPPortIn)
	updateChannel<- localnet.NextNode()
	fmt.Println(localnet.IsStartNode())
	for {
		select {
			case IPPing := <-UDPChan:
				time.Sleep(time.Millisecond)
				fmt.Println(IPPing)
				//localnet.PeerUpdate(IPPing)
			default:
				time.Sleep(time.Millisecond)
				/*
				if localnet.RemoveDeadConns(UDPTimeout) == true {
					updateChannel<- localnet.NextNode()
					if localnet.NextNode()==0{
						updateChannel<- "start"					
					} else {
						updateChannel<- "wait"
					}				
				}*/
		}
	}
}

func UDPReceiver(UDPReceiveChan chan string, passcode string, port string){	
	localIP, _ := localnet.IP()
	addr, _ := net.ResolveUDPAddr("udp", port)
	socket, _ := net.ListenUDP("udp", addr)
	//checkError(err, "Setting up UDP listener", iferror.Quit)

	reader := bufio.NewReader(socket)
	
	for {
		code, _ := reader.ReadString('\n')
		//checkError(err, "UDP datagram received", iferror.Ignore) // very frequent
		if code == (passcode + "\n") {
			msg, _ := reader.ReadString('\n')
			//fmt.Println(msg)
			// ignore computer's own messages
			if msg != (localIP + "\n"){	
				UDPReceiveChan <- msg[:len(msg)-1]
			}		
		} else {
			reader = bufio.NewReader(socket)
		}
	}
}

func UDPBroadcaster(channel chan bool, msg string, localBroadcastIP string, UDPport string){
		
	address, _ := net.ResolveUDPAddr("udp",localBroadcastIP+UDPport)
	conn, _ := net.DialUDP("udp", nil, address)
	//checkError(err, "Initialising UDP broadcast", iferror.Ignore)
	
	for {
		select {
			case done := <-channel:
				if done {
					return
				}
			default:
				_, _ = conn.Write( []byte(msg) )
				//checkError(err, "Broadcasting IP", iferror.Ignore)
				time.Sleep(time.Second)
		}
	}
}

