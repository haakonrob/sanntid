package main

import (
	"./network"
	"fmt"
	"strings"
	"time"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(20)
	incomingCh := make(chan interface{}, 1)
	outgoingCh := make(chan interface{}, 1)
	networkCh := make(chan string, 1)

	var online bool = false
	var localID string
	var timestamp = time.Now()
	activeElevs := make([]string, 0, 10)

	go network.Monitor(networkCh, true, "localhost", incomingCh, outgoingCh)
	for {
		select {
		case msg := <-networkCh:
			online, localID, activeElevs = decodeNetworkStatus(msg)
			fmt.Println("Net status: ", online, localID, activeElevs)
			/*
			if activeElevs[0] == localID {
				fmt.Println("Got here")
				outgoingCh <- "hello"
			}*/

		case msg := <-incomingCh:
			fmt.Println("Received: ", msg)
			//time.Sleep(time.Second)
			//outgoingCh <- msg
		default:
			if time.Since(timestamp) > time.Second*2 {
				timestamp = time.Now()
				outgoingCh <- "msg"
			}
			continue
		}
	}
}

func decodeNetworkStatus(str string) (bool, string, []string) {
	status := strings.Split(str, "_")
	online := false
	if status[0] == "true" {
		online = true
	}
	ID := status[1]
	list := strings.Split(status[2], "-")
	return online, ID, list
}
