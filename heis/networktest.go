package main

import (
	"fmt"
	"./networkmodule/networkmonitor"
)

func main(){
	nodeChan := make(chan string)
	monitorChan := make(chan string)
	networkmonitor.NetworkMonitor(nodeChan, monitorChan)

	for {
		select {
			case msg := <- nodeChan:
				fmt.Println(msg)
				nodeChan<- msg
			case msg := <- monitorChan:
				fmt.Println(msg)
		}
	}
}
