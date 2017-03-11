package main

import (
	"./network"
	"fmt"
	"runtime"
	"strings"
	"time"
)

func main() {
	runtime.GOMAXPROCS(20)
	incomingCh := make(chan string, 1)
	outgoingCh := make(chan string, 1)
	networkCh := make(chan string, 1)
	timestamp := time.Now()
	var online bool = false
	var localID string
	activeElevs := make([]string, 0, 10)
	go network.Monitor(networkCh, incomingCh, outgoingCh)
	for {
		select {
		case msg := <-networkCh:
			online, localID, activeElevs = decodeNetworkStatus(msg)
			fmt.Println("Update! Index: ", online, localID, activeElevs)

		case msg := <-incomingCh:
			fmt.Println("Received: ", msg)

		case msg := <-outgoingCh:
			fmt.Println(msg)

		default:
			if online && (time.Since(timestamp) > time.Second) {
				timestamp = time.Now()
				outgoingCh <- "hello"
				if "OK" == <-outgoingCh {
					fmt.Println("Sent!")
				} else {
					fmt.Println("Send failed.")
				}
			}
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
