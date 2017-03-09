package ring 

import (
	"net"
	"os"
	"fmt"
	_"json"
)

func NextNode(outgoingCh chan string, updateCh chan string, port int){
	initialised := false
	nextAddr := ""
	var conn net.Conn
	
	for {
		select {
		case nextAddr = <-updateCh:
			if initialised {
				conn.Close()
			}
			IP, _ := net.ResolveTCPAddr("tcp",nextAddr)
			conn, err = net.DialTCP("tcp", nil, IP)
			if err == nil {
				connUpdateCh<- "OK"
				initialised = true
			} else {
				connUpdateCh<- "ERROR"
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
	//var prevAddr = ""
	var conn net.Conn
	var buf [1024]byte
	
	for {
		if !initialised {
			go func(){
				ln, _ := net.ListenTCP("tcp", addr)
				conn, err := ln.Accept()
				if err != nil {
					initialised = true
				}
			}
		}
		select {
		case update := <-updateCh:
			if update == "RESET" {
				initialised = false
			}
		default:
			n, _, _ := conn.ReadFrom(buf[0:])
			msg := string(buf[:n])
			
		}
	}
}

}
