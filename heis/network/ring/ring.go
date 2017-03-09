package ring 

import (
	"net"
	"fmt"
	//"errors"
)

func NextNode(outgoingCh chan string, updateCh chan string, port int){
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
			IP, _ := net.ResolveTCPAddr("tcp",nextAddr)
			conn, err = net.DialTCP("tcp", nil, IP)
			if err == nil {
				updateCh<- "OK"
				initialised = true
			} else {
				updateCh<- "ERROR"
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

func PrevNode(incomingCh chan string, updateCh chan string, port int){
	initialised := false
	TCPAddr, _ := net.ResolveTCPAddr("tcp",fmt.Sprintf(":%d",port))
	var err error
	var conn net.Conn
	var buf [1024]byte
	
	for {
		if !initialised {
			ln, _ := net.ListenTCP("tcp", TCPAddr)
			conn, err = ln.Accept()
			if err != nil {
				initialised = true
			}
		}
		select {
		case update := <-updateCh:
			if update == "RESET" {
				initialised = false
			}
		default:
			n, err := conn.Read(buf[0:])
			if err != nil {
					initialised = false
					conn.Close()
			}
			msg := string(buf[:n])
			incomingCh<-msg
		}
	}
}

