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
				fmt.Println(msg)
				outgoingCh<- msg
				fmt.Println(<-outgoingCh)
				time.Sleep(time.Second)
			case msg := <- outgoingCh:
				fmt.Println(msg)
		}
	}
}
