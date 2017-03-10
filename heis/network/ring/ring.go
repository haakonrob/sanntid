package ring 

import (
	"net"
	"fmt"
	_"time"
	_"strings"
	//"errors"
)

func NextNode(outgoingCh chan string, updateCh chan string){
	initialised := false
	nextAddr := ""
	var conn net.Conn
	var err error
	for {
		select {
		case nextAddr = <-updateCh:
			if initialised {
				conn.Close()
			}
			/****TEMPORARY*****/
			//port = strings.Split(nextAddr, "-")[1]
			//nextAddr = strings.Split(nextAddr, "-")[0]
			/******************/
			IP, err := net.ResolveTCPAddr("tcp",nextAddr)
			if err != nil {
				fmt.Println("NextNode()",nextAddr)
			}
			conn, err = net.DialTCP("tcp", nil, IP)
			if err == nil {
				//updateCh<- "OK"
				initialised = true
				fmt.Println("ring.NextNode() OK")
			} else {
				fmt.Println("ring.NextNode() ERROR")
				//updateCh<- "ERROR"
				initialised = false	
			}
		case msg := <-outgoingCh:
			if initialised {
				_, err = conn.Write([]byte(msg))
			} else if (err != nil) || (!initialised) {
				outgoingCh<- "ERROR"
			} else {
				outgoingCh<- "OK"
			}
		}
	}

}

func PrevNode(incomingCh chan string, updateCh chan string, port string){
	var err error
	var conn net.Conn
	var buf [1024]byte

	initialised := false
	listening := false
	TCPAddr, err := net.ResolveTCPAddr("tcp", ":" + port)
	if err != nil {
		fmt.Println("PrevNode() Bad port")
	}
	
	for {
		if !initialised && !listening {
			go listenForTCP(TCPAddr, &initialised, &listening, &conn)
		}
		select {
		case update := <-updateCh:
			if update == "RESET" {
				fmt.Println("RESET")
				initialised = false
			}
		default:
			if initialised {
				fmt.Println("Trying to read")
				n, err := conn.Read(buf[0:])
				fmt.Println("Successfully read")
				if err != nil {
						initialised = false
						conn.Close()
				}else {
					msg := string(buf[:n])
					incomingCh<-msg
				}
			}
		}
	}
}

func listenForTCP( TCPAddr * net.TCPAddr, initialised * bool, listening * bool, conn *net.Conn)(){
	ln, err := net.ListenTCP("tcp", TCPAddr)
	if err != nil {
		*listening = false
		return
	}

	*conn, err = ln.Accept()
	if err == nil {
		*initialised = true
		fmt.Println("listenForTCP()",err)
	}
	*listening = false

}
