package main

import (
	"fmt"
	"./network"
)

func main(){
	
	incomingCh := make(chan string)
	outgoingCh := make(chan string)
	go network.Monitor(incomingCh, outgoingCh)
	for {
		select {
			case msg := <- incomingCh:
				fmt.Println(msg)
				outgoingCh<- msg
			case msg := <- outgoingCh:
				fmt.Println(msg)
		}
	}
}
