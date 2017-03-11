package main

import (
	"./network"
	"fmt"
	_ "runtime"
	"strings"
	"time"
)

func main() {

	incomingCh := make(chan string, 1)
	outgoingCh := make(chan string, 1)
	networkCh := make(chan string, 1)

	var online bool = false
	var localID string
	activeElevs := make([]string, 0, 10)

	go network.Monitor(networkCh, incomingCh, outgoingCh)

	for {
		select {
		case msg := <-networkCh:
			online, localID, activeElevs = decodeNetworkStatus(msg)
			fmt.Println("Net status: ", online, localID, activeElevs)
			if activeElevs[0] == localID {
				outgoingCh <- "hello"
			}

		case msg := <-incomingCh:
			fmt.Println("Received: ", msg)
			time.Sleep(time.Second)
			outgoingCh <- msg
		default:
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
