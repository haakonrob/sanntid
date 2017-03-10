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
	go network.Monitor(incomingCh, outgoingCh)
	for {
		select {
			case msg := <- incomingCh:
				fmt.Println("Received: ", msg)
				time.Sleep(time.Second)
				//outgoingCh<- msg
				//fmt.Println(<-outgoingCh)
				
			case msg := <- outgoingCh:
				fmt.Println(msg)
		}
	}
}
