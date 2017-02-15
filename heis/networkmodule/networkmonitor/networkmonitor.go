package networkmonitor
import (
	"net"
	"bufio"
	"time"
	"../localnet"
)


func JoinNetwork(elevChannel, nodeChannel, port, UDPPasscode){
	bcastMsg := UDPpasscode+"\n"+localnet.IP()+port+"\n"
	UDPChan := make(chan string)
	UDPBroadcastEnable := make(chan bool)
	go UDPReceiver(UDPChan, UDPpasscode, UDPport)
	go UDPBroadcaster(UDPBroadcastDone, UDPmsg, broadcastIP, UDPport)
}

func 

func UDPReceiver(UDPReceiveChan chan string, passcode string, port string){	
	addr, _ := net.ResolveUDPAddr("udp", port)
	socket, err := net.ListenUDP("udp", addr)
	//checkError(err, "Setting up UDP listener", iferror.Quit)

	reader := bufio.NewReader(socket)
	
	for {
		code, _ := reader.ReadString('\n')
		//checkError(err, "UDP datagram received", iferror.Ignore) // very frequent
		if code == (passcode + "\n") {
			msg, _ := reader.ReadString('\n')
			//fmt.Println(msg)
			// ignore computer's own messages
			if msg != (localnet.IP() + "\n"){	
				UDPReceiveChan <- msg[:len(msg)-1]
			}		
		} else {
			reader.Reset(socket)
		}
	}
}

func UDPBroadcaster(channel chan bool, msg string, localBroadcastIP string, UDPport string){
		
	address, err := net.ResolveUDPAddr("udp",localBroadcastIP+UDPport)
	conn, err := net.DialUDP("udp", nil, address)
	//checkError(err, "Initialising UDP broadcast", iferror.Ignore)
	
	for {
		select {
			case done = <-channel:
				if done {
					return
				}
			default:
				_, err = conn.Write( []byte(msg) )
				//checkError(err, "Broadcasting IP", iferror.Ignore)
				time.Sleep(time.Second)
		}
	}
}

