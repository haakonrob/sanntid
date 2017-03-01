package main

import (
	"fmt"
	"./networkmodule/networkmonitorTest"
)

func main(){
	nodeChan := make(chan string)
	monitorChan := make(chan string)
	networkmonitorTest.NetworkMonitor(nodeChan, monitorChan)

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
