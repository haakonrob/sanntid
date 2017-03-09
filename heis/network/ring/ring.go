package ring 

import (
	"os"
)

func Node(incomingCh chan string, outgoingCh chan string, connUpdateCh chan [2]string, port int){
	var initialised = false

	incomingChan := make(chan string)
	newIP := <-updateChannel
	go resetPrevNode(incomingChan, TCPPortIn)
	time.Sleep(time.Second)
	updateNextNode(newIP+TCPPortIn)
	
	
	for {
		select {
			case msg := <-connUpdateCh:
				switch(msg){
				case
				fmt.Println("Updated")
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


func NextNode(){}
