package main

import (
	"./network"
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(20)
	incomingCh := make(chan string)
	outgoingCh := make(chan string)
	networkCh := make(chan string)
	timestamp := time.Now()
	go network.Monitor(networkCh, incomingCh, outgoingCh)
	for {
		select {
		case msg := <-networkCh:

			fmt.Println("Update! Index: ", msg)

		case msg := <-incomingCh:
			fmt.Println("Received: ", msg)
		case msg := <-outgoingCh:
			fmt.Println(msg)

		default:
			if time.Since(timestamp) > time.Second {
				timestamp = time.Now()
				/*outgoingCh<- "hello"
				if "OK" != <-outgoingCh {
					fmt.Println("message failed")
				}*/
			}
		}
	}
}
