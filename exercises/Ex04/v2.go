package main 


import (
	"fmt"
	"net"
	"bufio"
	"time"
	"./localip"
	"./localnet"
	"./iferror"
)

const (
	TCPportIn = ":20024"
	TCPportOut = ":20025"
	UDPport = ":20023"
	UDPpasscode = "svekonrules"
)

type elevdata struct {
	externalOrders [6] int
	internalOrders [4] int
}

type elevPacket struct {
	msgType string
	msg string
	data elevdata
	
}

func main(){
	TCPNodeOnline := false
	localIP, err := localnet.GetLocalIP()
	broadcastIP, _ := localnet.GetBroadcastIP()
	checkError(err, "Retrieving local IP", iferror.Ignore)
	UDPmsg := UDPpasscode+"\n"+localIP+"\n"
	
	UDPChan := make(chan string)
	UDPBroadcastEnable := make(chan bool)
	TCPChan := make(chan string)
	stopTCP := make(chan bool)
	
	go UDPReceiver(UDPChan, UDPpasscode, UDPport)
	go UDPBroadcaster(UDPBroadcastEnable, UDPmsg, broadcastIP, UDPport)

	UDPBroadcastEnable <- true

	for {	
		select {
			case newIP := <-UDPChan:
				_ = localnet.AddNewNodeIP(newIP)
				if TCPNodeOnline {
					stopTCP<- true
				}
				go TCPNode(TCPChan, stopTCP, TCPportIn, TCPportOut)
				TCPNodeOnline = true
			case packet := <-TCPChan:
				fmt.Println(packet)
				time.Sleep(time.Second/2)
				TCPChan<- packet
			default:
				continue
		}
	}
}

func UDPReceiver(UDPReceiveChan chan string, passcode string, port string){
	localIP, _ := localip.Get()	
	addr, _ := net.ResolveUDPAddr("udp", port)
	socket, err := net.ListenUDP("udp", addr)
	checkError(err, "Setting up UDP listener", iferror.Quit)

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
			reader.Reset(socket)
		}
	}

}

func UDPBroadcaster(TransmitEnable chan bool, msg string, localBroadcastIP string, UDPport string){
		
	address, err := net.ResolveUDPAddr("udp",localBroadcastIP+UDPport)
	conn, err := net.DialUDP("udp", nil, address)
	checkError(err, "Initialising UDP broadcast", iferror.Ignore)
	
	broadcastOn := false
	for {
		// If off, wait for enable so that resources aren't taken.
		if !broadcastOn {
			broadcastOn = <-TransmitEnable		
		} else {
		// If on, check channel for updates and transmit every second.
			select {
				case broadcastOn = <-TransmitEnable:
					continue
				default:
					_, err = conn.Write( []byte(msg) )
					checkError(err, "Broadcasting IP", iferror.Ignore)
					time.Sleep(time.Second)
			}
		}
	}
}

func TCPNode(TCPChan chan string, stopTCP chan bool, portIn string, portOut string){
	if localnet.GetNumberOfNodes() == 2 {
		TCPNodePair(TCPChan, stopTCP, portIn)
	} else if localnet.GetNumberOfNodes() > 2 {
		TCPNodeRing(TCPChan, stopTCP, portIn, portOut)
	}
	
	
}
func TCPNodePair(TCPChan chan string, stopTCP chan bool, portIn string){
	var pairConn net.Conn
	buf := make([]byte, 1024)
	n := 0

	listen := func(){
		n = 0
		buf = make([]byte, 1024)
		n, _ = pairConn.Read(buf)
		return
	}	

	if localnet.IsStartNode() {
		pairConn, _ = setUpTCPConn(localnet.GetNextNodeIP(), portIn)
		TCPChan<- "startup"	
		pairConn.Write( []byte(<-TCPChan) )
	} else {
		pairConn, _ = listenForTCPConn(portIn)
	}
	
	go listen()
	for {
		select {
			case stop := <-stopTCP:
				if stop {
					pairConn.Close()
					return			
				}			
			default:
				// Read the nonempty buffer, wait for a reply, forward the reply, and restart the listener.
				if n !=0 {
					TCPChan<- string(buf)
					pairConn.Write( []byte(<-TCPChan) )
					go listen()
				}
		}
	}
	


}

func TCPNodeRing(TCPChan chan string, stopTCP chan bool, portIn string, portOut string){
	var nextNodeConn, prevNodeConn net.Conn
	buf := make([]byte, 1024)
	n := 0

	listen := func(){
		n = 0
		buf = make([]byte, 1024)
		n, _ = prevNodeConn.Read(buf)
		return
	}

	if localnet.IsStartNode() {
		nextNodeConn, _ = setUpTCPConn(localnet.GetNextNodeIP(), portOut)
		prevNodeConn, _ = listenForTCPConn(portIn)
	} else {
		prevNodeConn, _ = listenForTCPConn(portIn)
		nextNodeConn, _ = setUpTCPConn(localnet.GetNextNodeIP(), portOut)
	}
	
	// Start listening on lowerNodeConn
	go listen()
	
	if localnet.IsStartNode() {
		TCPChan<- "hello"
	}
	for {
		select {
			case stop := <-stopTCP:
				if stop {
					nextNodeConn.Close()
					prevNodeConn.Close() 
					return			
				}			
			default:
				// Read the nonempty buffer, wait for a reply, forward the reply, and restart the listener.
				if n !=0 {
					TCPChan<- string(buf)
					nextNodeConn.Write( []byte(<-TCPChan) )
					go listen()
				}
		}
	}
}

func listenForTCPConn(port string)(net.Conn, error){
	localAddress, err := net.ResolveTCPAddr("tcp", port)
	checkError(err, "Resolving local TCP address", iferror.Ignore)	
	listener, err := net.ListenTCP("tcp", localAddress)
	checkError(err, "Creating a TCP listener", iferror.Ignore)
	return listener.Accept()
}

func setUpTCPConn(targetIP string, port string)(net.Conn, error){
	upperNodeAddr, err := net.ResolveTCPAddr("tcp",targetIP+port)
	checkError(err, "Resolving next node IP", iferror.Ignore)	
	return net.DialTCP("tcp", nil, upperNodeAddr)
}

func checkError(err error, msg string, f iferror.Action){
	if err != nil {
		fmt.Println(msg, "... ERROR")
		fmt.Println(err)
		f()
	} else {
		fmt.Println(msg, "... Done")
	}
}

