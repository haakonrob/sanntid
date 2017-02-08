package main 


import (
	"fmt"
	"net"
	"bufio"
	"time"
	"./localip"
	"./iferror"
)

type elevPacket struct {
	
}

func main(){
	UDPport := ":20022"
	localIP, err := localip.Get()
	//localBroadcastIP, _ := localip.GetBroadcast()
	passcode := "svekonrules"
	UDPmsg := passcode+"\n"+localIP+"\n"
	checkError(err, "Retrieving local IP", iferror.Ignore)

	UDPReceiveChan := make(chan string)
	UDPTransmitEnable := make(chan bool)
	//TCPReceiveChan := make(chan elevPacket)
	//TCPTransmitChan := make(chan elevPacket)
	
	go UDPReceiver(UDPReceiveChan, passcode, UDPport)
	go UDPBroadcaster(UDPTransmitEnable, UDPmsg, localIP, UDPport)
	UDPTransmitEnable <- true
	for{
		fmt.Println("main",<-UDPReceiveChan)
		UDPTransmitEnable <- false
	}
}

func UDPReceiver(UDPReceiveChan chan string, passcode string, port string){
	localIP, _ := localip.Get()	
	addr, _ := net.ResolveUDPAddr("udp", port)
	socket, err := net.ListenUDP("udp", addr)
	checkError(err, "Setting up UDP listener", iferror.Quit)

	reader := bufio.NewReader(socket)
	
	for {
		code, err := reader.ReadString('\n')
		//checkError(err, "UDP datagram received", iferror.Ignore)
		if code == (passcode + "\n") {
			msg, _ := reader.ReadString('\n')
			//fmt.Println(msg)
			// ignore computer's own messages
			if msg != (localIP + "\n") || true {	
				UDPReceiveChan <- msg[:len(msg)-1]
			}		
		} else {
			err = nil
			for err == nil {
				_, err = reader.ReadString('\n')	
			}	
		}
	}

}

func UDPBroadcaster(TransmitEnable chan bool, msg string, localBroadcastIP string, UDPport string){
		
	address, err := net.ResolveUDPAddr("udp",localBroadcastIP+UDPport)
	conn, err := net.DialUDP("udp", nil, address)
	checkError(err, "Initialising UDP broadcast", iferror.Ignore)
	
	bytemsg := []byte(msg)
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
					_, err = conn.Write(bytemsg)
					//fmt.Println(msg)
					checkError(err, "Broadcasting IP", iferror.Ignore)
					time.Sleep(time.Second)
			}
		}
	}
}

func TCPReceiver(ReceiveChan chan elevPacket){}

func TCPTransmitter(ReceiveChan chan elevPacket){}



func checkError(err error, msg string, f iferror.Action){
	if err != nil {
		fmt.Println(msg, "... ERROR")
		fmt.Println(err)
		f()
	} else {
		fmt.Println(msg, "... Done")
	}
}
