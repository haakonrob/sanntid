package main

import (
	"./network"
	"fmt"
	"strings"
	"time"
)

func main() {
	netCh := make(chan string)
	incomingCh := make(chan string)
	outgoingCh := make(chan string)
	timestamp := time.Now()
	ringMsg := "hello"

	var online bool
	var localID string
	var activePeers []string

	go network.Monitor(netCh, incomingCh, outgoingCh)

	for {
		select {
		case msg := <-netCh:
			online, localID, activePeers = decodeNetworkStatus(msg)
			fmt.Println("Status: ", online, localID, activePeers)
		case msg := <-incomingCh:
			fmt.Println(msg)
		default:
			if online && (time.Since(timestamp) > time.Second) {
				timestamp = time.Now()
				outgoingCh <- ringMsg
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
