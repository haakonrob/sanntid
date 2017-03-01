package ringnode

import (
	"net"
	"time"
)

var (
	nextNode net.Conn
	prevNode net.Conn
)

func RingNode(elevChannel chan string, updateChannel chan string, TCPPortIn string, TCPPortOut string){
	incomingChan := make(chan string)
	msg := <-updateChannel
	updateNextNode(msg,TCPPortIn)
	go resetPrevNode(incomingChan, TCPPortIn)
	for {
		select {
			case msg := <-updateChannel:
				go resetPrevNode(incomingChan, TCPPortIn)
				updateNextNode(msg,TCPPortOut)
			case msg := <-incomingChan:
				elevChannel<- msg
				time.Sleep(time.Second)
				nextNode.Write([]byte(msg)) 
				// errorCheck
				//go resetPrevNode(incomingChan, TCPPortIn)
		}
	}
}
/*
func prevNodeListen(incomingChan chan string){
	buf := make([]byte, 1024)
	n, _ := prevNode.Read(buf)
	//CheckError(err, " ")
	//fmt.Println("Received: ", string(buf[0:n]), "\n")		
	//fmt.Println(err)
	
}*/

func resetPrevNode(incomingChan chan string, port string){
	prevNode.Close()
	addr, _ := net.ResolveTCPAddr("tcp", port)
	listener, _ := net.ListenTCP("tcp", addr)
	prevNode, _ = listener.Accept()
	buf := make([]byte, 1024)
	n, _ := prevNode.Read(buf)
	incomingChan<- string(buf[0:n])
	return
}

func updateNextNode(nextNodeIP string, port string){
	nextNode.Close()
	address, _ := net.ResolveTCPAddr("tcp",nextNodeIP+port)
	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		//return
	}
	nextNode = conn
}

