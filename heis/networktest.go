package main

import (
	"fmt"
	"time"
	"./network"
	"runtime"
)

func main(){
	runtime.GOMAXPROCS(20)
	incomingCh := make(chan string)
	outgoingCh := make(chan string)
	networkCh := make(chan string)

	go network.Monitor(networkCh, incomingCh, outgoingCh)
	for {
		select {
		case msg := <-networkCh:
			fmt.Println("Update! Index: ", msg)
			if msg == "0"{
				outgoingCh<- "hello"
			}
		case msg := <- incomingCh:
			fmt.Println("Received: ", msg)
			time.Sleep(time.Second)
			outgoingCh<- msg
			if "OK" != <-outgoingCh {
				fmt.Println("message failed")
			}
			
				
		case msg := <- outgoingCh:
			fmt.Println(msg)
		}
	}
}
