package ringnode

import (
	"net"
	"time"
	"fmt"
)

var (
	nextNode net.Conn
	prevNode net.Conn
)

func RingNode(elevChannel chan string, updateChannel chan string, TCPPortIn string, TCPPortOut string){
	incomingChan := make(chan string)
	newIP := <-updateChannel
	go resetPrevNode(incomingChan, TCPPortIn)
	time.Sleep(time.Second)
	updateNextNode(newIP)
	fmt.Println("Updated")
	
	for {
		select {
			case msg := <-updateChannel:
				go resetPrevNode(incomingChan, TCPPortIn)
				updateNextNode(msg)
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
	//prevNode.Close()
	addr, _ := net.ResolveTCPAddr("tcp", port)
	listener, _ := net.ListenTCP("tcp", addr)
	prevNode, _ = listener.Accept()
	fmt.Println("PrevConnect")
	buf := make([]byte, 1024)
	n, _ := prevNode.Read(buf)
	incomingChan<- string(buf[0:n])
	return
}

func updateNextNode(nextNodeIP string){
	//nextNode.Close()
	address, err := net.ResolveTCPAddr("tcp",nextNodeIP)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		fmt.Println(err)
		return
	}
	nextNode = conn
}

