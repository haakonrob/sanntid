package main 


import (
	"fmt"
	"net"
	"bufio"
	"time"
	"sort"
	"./localip"
	"./iferror"
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
	isStartNode := false
	UDPport := ":20023"
	TCPportIn := ":20024"
	TCPportOut := ":20024"
	localIP, err := localip.Get()
	broadcastIP, _ := localip.GetBroadcast()
	checkError(err, "Retrieving local IP", iferror.Ignore)
	passcode := "svekonrules"
	UDPmsg := passcode+"\n"+localIP+"\n"
	IPList := make([] string, 0, 20)
	
	UDPReceiveChan := make(chan string)
	UDPTransmitEnable := make(chan bool)
	TCPChan := make(chan string)
	stopTCP := make(chan bool)
	
	go UDPReceiver(UDPReceiveChan, passcode, UDPport)
	go UDPBroadcaster(UDPTransmitEnable, UDPmsg, broadcastIP, UDPport)
	UDPTransmitEnable <- true

	for {	
		select {
			case newIP := <-UDPReceiveChan:
				if dontKnowIP(newIP, IPList){
					IPList = append(IPList, newIP)
					sort.Strings(IPList)
					fmt.Println(IPList)
					if TCPNodeOnline {
						stopTCP<- true
					}
					go TCPNode(TCPChan, stopTCP, TCPportIn, TCPportOut, IPList)
					
				} else {
					// try to reconnect with failing node? Perhaps count the number of times until a restart seems smart.
				}
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


func TCPNode(TCPChan chan string, stopTCP chan bool, portIn string, portOut string, IPList []string){
	numberOfNodes := len(IPList)+1
	localIP := localip.Get()
	var nextNodeIP string
	if localIP < IPList[0] {
		isStartNode = true
	}
	if numberOfNodes == 2 {
		nextNodeIP = IPList[0]
	} else if numberOfNodes > 2 {
		nextNodeIP = getUpperNode(IPList)
	}	
	

	localAddress, err := net.ResolveTCPAddr("tcp", portIn)
	checkError(err, "Resolving local TCP address", iferror.Ignore)	
	
	listener, err := net.ListenTCP("tcp", localAddress)
	checkError(err, "Creating a TCP listener", iferror.Ignore)

	upperNodeAddr, err := net.ResolveTCPAddr("tcp",nextNodeIP+portOut)
	checkError(err, "Resolving next node IP", iferror.Quit)
	fmt.Println("alive")
	var lowerNodeConn net.Conn
	var upperNodeConn net.Conn
	fmt.Println("alive")
	if isStartNode {
		upperNodeConn, _ = net.DialTCP("tcp", nil, upperNodeAddr)
		lowerNodeConn, _ = listener.Accept()
	} else {
		lowerNodeConn, _ = listener.Accept()
		upperNodeConn, _ = net.DialTCP("tcp", nil, upperNodeAddr)
	}
	fmt.Println("alive")
	buf := make([]byte, 1024)
	n := 0

	listen := func(){
		n = 0
		buf = make([]byte, 1024)
		n,err = lowerNodeConn.Read(buf)
		return
	}	
	
	// Start listening on lowerNodeConn
	go listen()
	
	if isStartNode {
		TCPChan<- "hello"
	}
	for {
		select {
			case stop := <-stopTCP:
				if stop {
					upperNodeConn.Close()
					lowerNodeConn.Close() 
					return			
				}			
			default:
				// Read the nonempty buffer, wait for a reply, forward the reply, and restart the listener.
				if n !=0 {
					TCPChan<- string(buf)
					upperNodeConn.Write( []byte(<-TCPChan) )
					go listen()
				}
		}
	}
}

func dontKnowIP(IP string, IPlist []string)(bool){
	for i:=0; i<len(IPlist); i++ {
		if IPlist[i] == IP {
			return false
		} 
	}
	return true
}

func getUpperNode(IPList []string)(string){
	localIP, _ := localip.Get()
	
	// smallest member
	if localIP < IPList[0] {
		return IPList[0]
	}
	// somewhere inbetween
	for i:=0;i<len(IPList)-1;i++ {
		if localIP == IPList[i] {
			return "BadIPlist" //shouldn't happen
		} else if localIP > IPList[i] && localIP < IPList[i+1] {
			return IPList[i+1]
		} 
	}
	// reached end of list, wrap around
	return IPList[0]
	
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

