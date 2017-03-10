package ring 

import (
	"net"
	"fmt"
	"time"
	"strings"
)

const checkInterval = 500*time.Millisecond

func NextNode(outgoingCh chan string, updateCh chan string){
	initialised := false
	nextAddr := ""
	var conn net.Conn
	var err error
	var lastConnCheck = time.Now().Add(time.Millisecond)

	for {
		select {
		case nextAddr = <-updateCh:
			if initialised {
				conn.Close()
			}
			IP, err := net.ResolveTCPAddr("tcp",nextAddr)
			if err != nil {
				fmt.Println("NextNode()",nextAddr)
			}
			conn, err = net.DialTCP("tcp", nil, IP)
			if err == nil {
				updateCh<- "OK"
				initialised = true
				//fmt.Println("ring.NextNode()", IP)
			} else {
				fmt.Println("ring.NextNode() ERROR, tried ", nextAddr)
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
		default:
			if time.Since(lastConnCheck) > checkInterval && initialised {
				_, err = conn.Write([]byte("heartbeat\n"))
				if (err != nil) {
					conn.Close()
					initialised = false
					updateCh<- "ERROR"
				}
				/*buf := []byte{}
				conn.SetReadDeadline(time.Now().Add(time.Millisecond))
				if _, err := conn.Read(buf); err != nil {
					
				}*/
			}
		}
	}
}

func PrevNode(incomingCh chan string, updateCh chan string, port int){
	var err error
	var conn net.Conn
	var buf [1024]byte

	initialised := false
	listening := false
	TCPAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("PrevNode() Bad port: ", fmt.Sprintf(":%d", port))
	}
	
	for {
		if !initialised && !listening {
			listening = true
			go listenForTCP(TCPAddr, &initialised, &listening, &conn)
		}
		select {
		case update := <-updateCh:
			if update == "RESET" {
				//fmt.Println("RESET")
				initialised = false
			}
		default:
			if initialised {
				//fmt.Println("Trying to read")
				n, err := conn.Read(buf[0:])
				//fmt.Println("Successfully read")
				if err == nil {
					msg := string(buf[:n])
					if !strings.Contains(msg, "heartbeat") {
						incomingCh<-msg	
					}
				} else {
					fmt.Println("Failed to read prevNode")
					initialised = false
					conn.Close()
					
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
		fmt.Println("prevNode OK")
	}
	*listening = false

}
