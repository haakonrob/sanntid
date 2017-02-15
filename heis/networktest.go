package main

import (
	"fmt"
	"./networkmodule"
)

func main(){
	nodeChan := make(chan string)
	monitorChan := make(chan string)
	networkmodule.Init(nodeChan, monitorChan)

	for {
		select {
			case msg := <- nodeChan:
				fmt.Println(msg)
			case msg := <- monitorChan:
				fmt.Println(msg)
		}
	}
}
