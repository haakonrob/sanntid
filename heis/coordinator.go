package main

import (
	"./backup"
	heis "./heisdriver" //"./simulator/client"
	"./network"
	"./operator"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

/******************************************
The
******************************************/

const (
	MAX_NUM_ELEVS = 10
	N_FLOORS      = heis.N_FLOORS
	UP            = heis.BUTTON_CALL_UP
	DOWN          = heis.BUTTON_CALL_DOWN
	COMMAND       = heis.BUTTON_COMMAND
	loopBack      = false
	subnet        = "sanntidsal"
)

type Message struct {
	SenderID  string                      //'json:"SenderID"'
	ViewCount int                         //'json:"ViewCount"'
	Scores    map[string][2][N_FLOORS]int //'json:"Scores"'
	Content   GlobalOrderStruct           //'json:"Content"'
}

type GlobalOrderStruct struct {
	Available  [2][N_FLOORS]bool      //'json:"Available"'
	Taken      [2][N_FLOORS]bool      //'json:"Taken"'
	Timestamps [2][N_FLOORS]time.Time //'json:"Timestamps"'
}

var (
	GlobalMask   GlobalOrderStruct
	GlobalOrders GlobalOrderStruct
	LocalOrders  operator.Orders
	online       bool
	localID      string
	activeElevs  []string
	updateFlag   bool
)

/*********************************
Main

*********************************/

func main() {
	activeElevs = make([]string, 0, MAX_NUM_ELEVS)

	orderChan := make(chan heis.Order, 5)
	completedOrdersCh := make(chan operator.Orders, 1)
	eventChan := make(chan heis.Event, 5)
	operatorChan := make(chan operator.Orders, 1)
	networkCh := make(chan network.Status)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	timeoutCh := make(chan heis.Order)

	heis.ElevInit()
	go operator.Start(eventChan, operatorChan, completedOrdersCh)
	go network.Monitor(networkCh, loopBack, "sanntidsal", incomingCh, outgoingCh)
	go heis.Poller(orderChan, eventChan)
	//go timeOut(timeoutCh)

	if len(os.Args) > 1 {
		if temp, ok := backup.Load(os.Args[1]); ok {
			LocalOrders = temp.(operator.Orders)
		}

		//Avoid timeouting again on restart.
		for o := UP; o <= COMMAND; o++ {
			for f := 0; f < N_FLOORS; f++ {
				LocalOrders.Timestamps[o][f] = time.Now()
			}
		}
	}

	tickChan := time.NewTicker(time.Millisecond * 1000).C

	resetMask()

	for {

		select {
		case status := <-networkCh:
			online, localID, activeElevs = status.Online, status.LocalID, status.ActiveIDs
			updateFlag = true

		case newOrder := <-orderChan:
			markAsAvailable(newOrder)

		case completedOrders := <-completedOrdersCh:
			markAsCompleted(completedOrders)

		case jmsg := <-incomingCh:
			msg := Message{}
			_ = json.Unmarshal(jmsg, &msg)

			switch msg.ViewCount {
			case 0:
				applyMaskToGlobalOrders(msg.Content)
				msg.Scores[localID] = [2][N_FLOORS]int{}
				msg.Scores[localID] = scoreOrders()

				if msg.SenderID == localID {
					msg.ViewCount++
				}
				jmsg, _ := json.Marshal(msg)
				outgoingCh <- jmsg

			case 1:
				getLocalOrders(msg.Scores)
				if msg.SenderID != localID {
					jmsg, _ := json.Marshal(msg)
					outgoingCh <- jmsg
				}
			default:
				fmt.Println("Received invalid message")
			}

		case <-tickChan:
			printOrders()

			//if updateFlag {
			fmt.Println("Sending")
			switch online {
			case true:
				msg := Message{
					SenderID:  localID,
					ViewCount: 0,
					Scores:    map[string][2][N_FLOORS]int{},
					Content:   GlobalMask,
				}
				jmsg, _ := json.Marshal(msg)
				outgoingCh <- jmsg

			case false:
				applyMaskToGlobalOrders(GlobalMask)
				takeAllAvailableOrders()
			}
			//updateFlag = false
			backup.Write(LocalOrders)
			setLights()
			resetMask()
			operatorChan <- LocalOrders

		case lateOrder := <-timeoutCh:
			GlobalOrders.Taken[lateOrder.OrderType][lateOrder.Floor] = false
			//markAsAvailable(lateOrder)
			if LocalOrders.Pending[lateOrder.OrderType][lateOrder.Floor] {
				//backup.Restart()
			}
		}
	}
}

/*************************************************************************************************
 Functions for handling and merging LocalOrders and GlobalOrders.
 Most of the complexity is due to the array iteration.

*************************************************************************************************/

func setLights() {
	for o := heis.BUTTON_CALL_UP; o < heis.BUTTON_COMMAND; o++ {
		for f := 0; f < N_FLOORS; f++ {
			val := GlobalOrders.Taken[o][f]
			heis.ElevSetButtonLamp(o, f, val)
		}
	}
	for f := 0; f < N_FLOORS; f++ {
		heis.ElevSetButtonLamp(heis.BUTTON_COMMAND, f, LocalOrders.Pending[2][f])
	}
}

func resetMask() {
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			GlobalMask.Available[o][f] = false
			GlobalMask.Taken[o][f] = true
		}
	}
}

func applyMaskToGlobalOrders(mask GlobalOrderStruct) {
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			GlobalOrders.Available[o][f] = mask.Available[o][f] || GlobalOrders.Available[o][f]
			GlobalOrders.Taken[o][f] = mask.Taken[o][f] && GlobalOrders.Taken[o][f]
		}
	}
}

func markAsAvailable(order heis.Order) {
	o, f := order.OrderType, order.Floor
	switch o {
	case UP, DOWN:
		if GlobalMask.Taken[o][f] && !GlobalOrders.Taken[o][f] {
			GlobalMask.Available[o][f] = true
			GlobalMask.Timestamps[o][f] = time.Now()
			updateFlag = true
		}
	case COMMAND:
		if !LocalOrders.Completed[o][f] {
			LocalOrders.Pending[o][f] = true
			LocalOrders.Timestamps[o][f] = time.Now()
			updateFlag = true
		}
	default:
		fmt.Println("Cannot add an invalid order.")
	}
}

func markAsCompleted(completedOrders operator.Orders) {
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			if completedOrders.Completed[o][f] {
				GlobalMask.Taken[o][f] = false
				LocalOrders.Pending[o][f] = false
			}
		}
	}
}

func scoreOrders() [2][N_FLOORS]int {
	scores := [2][N_FLOORS]int{}
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			floorDiff := (operator.GetPrevFloor() - f)
			if floorDiff < 0 {
				floorDiff = -floorDiff
			}
			scores[o][f] = 200 - floorDiff*40
		}
	}
	return scores
}

func takeAllAvailableOrders() {
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			if GlobalOrders.Available[o][f] {
				GlobalOrders.Available[o][f] = false
				GlobalOrders.Taken[o][f] = true
				LocalOrders.Pending[o][f] = true
			}
		}
	}
}

func getLocalOrders(scores map[string][2][N_FLOORS]int) {
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			if GlobalOrders.Available[o][f] {
				maxID, max := activeElevs[0], 0
				for _, id := range activeElevs {
					if val, ok := scores[id]; ok {
						if val[o][f] > max {
							max = val[o][f]
							maxID = id
						}
					}
				}
				if maxID == localID {
					GlobalOrders.Available[o][f] = false
					GlobalOrders.Taken[o][f] = true
					LocalOrders.Pending[o][f] = true
				} else {
					GlobalOrders.Available[o][f] = false
					GlobalOrders.Taken[o][f] = true
				}
			}
		}
	}
}

func timeOut(timeoutCh chan heis.Order) {
	time.Sleep(time.Second)
	for o := UP; o < COMMAND; o++ {
		for f := 0; f < N_FLOORS; f++ {
			if LocalOrders.Pending[o][f] {
				if time.Since(LocalOrders.Timestamps[o][f]) > time.Second*15 {
					timeoutCh <- heis.Order{o, f}
				}
			}
			if GlobalOrders.Taken[o][f] {
				if time.Since(GlobalOrders.Timestamps[o][f]) > time.Second*15 {
					timeoutCh <- heis.Order{o, f}
				}
			}
		}
	}
}

func printOrders() {
	fmt.Printf("\033[0;0H")
	for i := 0; i < 50; i++ {
		for j := 0; j < 100; j++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n\n\n\n\n\n")
	fmt.Println("Updated:  ", updateFlag)
	fmt.Println("Online:  ", online)
	fmt.Println("localID:  ", localID)
	fmt.Println("Active IDs:                      ")
	fmt.Println(activeElevs)
	fmt.Println("--------------------------------------------------------")
	fmt.Println("Available orders:                ")
	fmt.Println("UP:    ", GlobalOrders.Available[0])
	fmt.Println("DOWN:  ", GlobalOrders.Available[1])
	fmt.Println("--------------------------------------------------------")
	fmt.Println("Taken Orders:                    ")
	fmt.Println("UP:    ", GlobalOrders.Taken[0])
	fmt.Println("DOWN:  ", GlobalOrders.Taken[1])
	fmt.Println("--------------------------------------------------------")
	fmt.Println("Scores:")
	/*
		for _, elevID := range activeElevs {
			if scores, ok := GlobalOrders.Scores[elevID]; ok {
				fmt.Println("ID:  ", elevID)
				fmt.Println("UP:    ", scores[0])
				fmt.Println("DOWN:  ", scores[1])
			}
		}
	*/
	fmt.Println("--------------------------------------------------------")
	fmt.Println("LocalOrders Pending:")
	fmt.Println("UP:    ", LocalOrders.Pending[0])
	fmt.Println("DOWN:  ", LocalOrders.Pending[1])
	fmt.Println("COMM:  ", LocalOrders.Pending[2])
	fmt.Println("--------------------------------------------------------")
	fmt.Println("--------------------------------------------------------")
	fmt.Println("LocalOrders Complete:")
	fmt.Println("UP:    ", LocalOrders.Completed[0])
	fmt.Println("DOWN:  ", LocalOrders.Completed[1])
	fmt.Println("COMM:  ", LocalOrders.Completed[2])
	fmt.Println("--------------------------------------------------------")
}
