package main

import (
	"./fsm"
	heis "./heisdriver" //"./simulator/client"
	"./network"
	"encoding/json"
	"fmt"
	_ "os"
	"strings"
	"time"
)

/******************************************
TODO - Functionality
Insert logic for network state ( monitor returns {online, ID, activeElevs} )
Cost function for scoring of orders
Better logic for taking orders - all active elevs must have a score in order to take an order, when taking, set all scores to -1.
Maybe redundant - add a way of seeing if all elevs have viewed the packet
Change global orders Scores and Backup to maps for use with elev IDs
change localElevIndex to ID string, returned from networkmonitor. OR set at start like in example code.

TODO - Fault tolerance
Add timestamping logic
Data logging for backup in case of crash/termination


DISCUSS
Timestamps in Taken instead of heisnr?
******************************************/

//elevtype heis.ElevType = heis.ET_Simulation

const (
	MAX_NUM_ELEVS = 10
	N_FLOORS      = heis.N_FLOORS
	UP            = heis.BUTTON_CALL_UP
	DOWN          = heis.BUTTON_CALL_DOWN
	COMMAND       = heis.BUTTON_COMMAND
)

type GlobalOrderStruct struct {
	Available  [2][N_FLOORS]bool      //'json:"Available"'
	Taken      [2][N_FLOORS]bool      //'json:"Taken"'
	Timestamps [2][N_FLOORS]time.Time //'json:"Timestamps"'
	Scores            map[string][2][N_FLOORS]int        //[MAX_NUM_ELEVS][2][N_FLOORS]int    //'json:"Scores"'
	LocalOrdersBackup [MAX_NUM_ELEVS]fsm.LocalOrderState //'json:"LocalOrdersBackup"'
	SenderId          string
} //

var (
	GlobalOrders GlobalOrderStruct
	LocalOrders fsm.LocalOrderState

	online      bool
	localID     string
	activeElevs []string

	orderTimestamp int
)

/*********************************
Testing for network encoding
*********************************/
var str string

/*********************************
Main
*********************************/

func main() {

	// Init stuff
	/************************/
	online = false
	activeElevs = make([]string, 0, MAX_NUM_ELEVS)

	// Sets all Taken values to -1. Not o because 0 can be a elevIndex
	for ordertype := UP; ordertype <= DOWN; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			GlobalOrders.Taken[ordertype][floor] = false
		}
	}
	/************************/

	orderChan := make(chan heis.Order, 5)
	eventChan := make(chan heis.Event, 5)
	fsmChan := make(chan fsm.LocalOrderState)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	networkCh := make(chan string)

	/*****ADD*****
	networkChan := make(chan string)
	**************/
	heis.ElevInit()
	go heis.Poller(orderChan, eventChan)
	go fsm.Fsm(eventChan, fsmChan)
	go network.Monitor(networkCh, incomingCh, outgoingCh)
	/*****ADD*****
	go networkMonitor(network)


	**************/
	//timestamp := time.Now()
	for {
		// online, localElevIndex,  = networkStatus()
		// update score map
		// elev1

		select {
		case newOrder := <-orderChan:
			fmt.Println("Coord new order:", newOrder)

			if !online || (newOrder.OrderType == COMMAND) {
				//update local orders
				updateLocalOrders(newOrder)

				time.Sleep(time.Millisecond * 5)
				fsmChan <- LocalOrders
			} else {
				//update global avaliable
				updateGlobalOrders(newOrder)
				GlobalOrders.SenderId = localID
				outgoingCh <- EncodeGlobalPacket()
				//save in buffer
			}
		case completedOrders := <-fsmChan:
			fmt.Println("Coord completed order")
			if !online {
				//update local complete
				updateLocalComplete(completedOrders)
				updateLocalState(completedOrders)

			} else {
				//update global complete
				updateGlobalComplete(completedOrders)
				GlobalOrders.SenderId = localID
				outgoingCh <- EncodeGlobalPacket()

				//save in bufferout on network

			}

		case status := <-networkCh:
			online, localID, activeElevs = decodeNetworkStatus(status)
		case msg := <-incomingCh:
			oldGlobalOrders := GlobalOrders
			DecodeGlobalPacket(msg)
			mergeOrders(oldGlobalOrders)
			//time.Sleep(time.Millisecond * 100)

			scoreOrders()

			if GlobalOrders.SenderId == localID {
				takeGlobalOrders()
			} else {
				outgoingCh <- EncodeGlobalPacket()
				fmt.Println("Global orders: ", GlobalOrders.Available)
			}

			//update global all

			/*********ADD*********

			//if all scores done && you best score
				//update global taken
				//timestamp
				//delete all global scores for order

				//update local orders from global

			//else
				updateLocalState()
				//set your score
				//update global scores

				//send msg*/

		default:
			//timeout error handeling
			//if global taken too long
			//move from taken to avaliable
			//

			/*for ordertype := UP; ordertype <= DOWN; ordertype++ {
				for floor := 0; floor < N_FLOORS; floor++ {
					if time.Since(GlobalOrders.Timestamps[ordertype][floor]) > time.Second*10 {
						GlobalOrders.Available[ordertype][floor] = true
						GlobalOrders.Taken[ordertype][floor] = false
					}

					if time.Since(LocalOrders.Timestamps[ordertype][floor]) > time.Second*10 {
						GlobalOrders.Available[ordertype][floor] = true
						GlobalOrders.Taken[ordertype][floor] = false
						LocalOrders.Pending[ordertype][floor] = false
					}
				}
			}

			for floor := 0; floor < N_FLOORS; floor++ {
				if time.Since(LocalOrders.Timestamp[COMMAND][floor]) > time.Second*10 {
					//handel it by restart of system
				}
			}*/

			continue
		}
	}
}

/*********************************
Functions
Needed:
updateGlobalState()

*********************************/

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

func updateLocalState(newLocalState fsm.LocalOrderState) {
	LocalOrders.PrevFloor = newLocalState.PrevFloor
	LocalOrders.Direction = newLocalState.Direction
}

/**************sverre lør***************************/
func updateLocalOrders(order heis.Order) {
	ordertype := order.OrderType
	floor := order.Floor
	stamp := time.Now()
	switch ordertype {

	case UP:
		LocalOrders.Pending[ordertype][floor] = true
		LocalOrders.Completed[ordertype][floor] = false
		LocalOrders.Timestamps[ordertype][floor] = stamp
	case DOWN:

		LocalOrders.Pending[ordertype][floor] = true
		LocalOrders.Completed[ordertype][floor] = false
		LocalOrders.Timestamps[ordertype][floor] = stamp

	case COMMAND:
		LocalOrders.Pending[ordertype][floor] = true
		LocalOrders.Completed[ordertype][floor] = false
		LocalOrders.Timestamps[ordertype][floor] = stamp

	default:
		fmt.Println("Invalid OrderType in updatelocalorders()")
	}

}

func updateLocalComplete(newLocalState fsm.LocalOrderState) {
	for ordertype := UP; ordertype <= COMMAND; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			LocalOrders.Completed[ordertype][floor] = newLocalState.Completed[ordertype][floor]
			//LocalOrders.Pending[ordertype][floor] = LocalOrders.Pending[ordertype][floor] && !(LocalOrders.Completed[ordertype][floor])
		}
	}
}

func updateGlobalOrders(order heis.Order) {
	ordertype := order.OrderType
	floor := order.Floor
	stamp := time.Now()
	switch ordertype {

	case UP:
		GlobalOrders.Available[ordertype][floor] = true
		GlobalOrders.Taken[ordertype][floor] = false
		GlobalOrders.Timestamps[ordertype][floor] = stamp

	case DOWN:

		GlobalOrders.Available[ordertype][floor] = true
		GlobalOrders.Taken[ordertype][floor] = false
		GlobalOrders.Timestamps[ordertype][floor] = stamp

	default:
		fmt.Println("Invalid OrderType in updateglobalorders()")
	}

}

func updateGlobalComplete(newLocalState fsm.LocalOrderState) {

	for ordertype := UP; ordertype <= DOWN; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			if !GlobalOrders.Taken[ordertype][floor] && LocalOrders.Completed[ordertype][floor] {
				//!GlobalOrders.Taken[ordertype][floor] && LocalOrders.Completed[ordertype][floor]
			}
		}
	}
}

/*************************************************/

func updateGlobalState() {
	//merge unvertified and global orders
	//GlobalOrders = unverified_GlobalOrders
	setLights()

}

func setLights() {
	for b := 0; b < heis.N_BUTTONS-1; b++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			if GlobalOrders.Available[b][floor] || GlobalOrders.Taken[b][floor] {
				heis.ElevSetButtonLamp(b, floor, 1)
			} else {
				heis.ElevSetButtonLamp(b, floor, 0)
			}

		}
	}
}

/*
// getNextOrder() will be replaced by updateLocalOrders() AND mergeOrders()()
func getNextOrder() (heis.ElevButtonType, int, bool) {
	//IMPROVEMENTS:
	//Collect and choose best option instead of taking first one

	// If elev score is best in table AND not 0, it claims the order
	for ordertype := UP; ordertype <= DOWN; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			if GlobalOrders.Available[ordertype][floor] {
				if isBestScore(ordertype, floor) {
					GlobalOrders.Available[ordertype][floor] = false
					GlobalOrders.Taken[ordertype][floor] = true
					//GlobalOrders.Timestamps[ordertype][floor] = GlobalOrders.Clock
					return ordertype, floor, true
				}
			}
		}
	}
	return -1, -1, false
}
*/

func scoreOrders() {
	var scoresTemp [2][N_FLOORS]int

	for ordertype := UP; ordertype <= DOWN; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			//INSERT COST FUNC HERE
			//simple:
			//GlobalOrders.Available[ordertype][floor]

			scoresTemp[ordertype][floor] = 10 //LocalOrders.Pending[ordertype][floor]*10;
		}
	}
	if _, ok := GlobalOrders.Scores[localID]; ok {
		GlobalOrders.Scores[localID] = scoresTemp

	}
}

func isBestScore(ordertype heis.ElevButtonType, floor int) bool {
	// returns false if it finds a better competitor, else returns true.
	// If this is somehow called when the network is down, ignore globalorders.
	if len(activeElevs) == 0 || !online {
		return true
	}

	for _, elevID := range activeElevs {
		if value, ok := GlobalOrders.Scores[elevID]; ok {
			if GlobalOrders.Scores[localID][ordertype][floor] < value[ordertype][floor] {
				return false
			} else if GlobalOrders.Scores[elevID][ordertype][floor] == 0 {
				return false
			}
		}

	}
	return true
}

func takeGlobalOrders() {

	for ordertype := UP; ordertype <= DOWN; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			if isBestScore(ordertype, floor) && GlobalOrders.Available[ordertype][floor] {
				GlobalOrders.Available[ordertype][floor] = false
				GlobalOrders.Taken[ordertype][floor] = true
				LocalOrders.Pending[ordertype][floor] = true

			}
		}
	}
}

func mergeOrders(newGlobalOrders GlobalOrderStruct) {
	GlobalOrders = newGlobalOrders
	//^uint(0) gives the maximum value for uint
	//GlobalOrders.Clock = (GlobalOrders.Clock + 1) % (^uint(0))
	// If an order is listed as taken, but this elev has completed it, the order is removed from globalorders
	for ordertype := UP; ordertype <= DOWN; ordertype++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			if newGlobalOrders.Taken[ordertype][floor] && LocalOrders.Completed[ordertype][floor] {
				GlobalOrders.Taken[ordertype][floor] = false
				LocalOrders.Completed[ordertype][floor] = false
			}
		}
	}
}

/*******TEST DECODING ENCODING PACKET*********
GlobalPacketENC := EncodeGlobalPacket()
//fmt.Println(string(GlobalPacketENC))
_ = GlobalPacketENC

GlobalPacketDEC, err := DecodeGlobalPacket(GlobalPacketENC)
fmt.Println("Test PacketDEC: ", GlobalPacketDEC.Taken)
_ = err
****************************************/

func EncodeGlobalPacket() (b []byte) {
	GlobalPacketD, err := json.Marshal(GlobalOrders)
	_ = err
	return GlobalPacketD
}

func DecodeGlobalPacket(JsonPacket []byte) (PacketDEC GlobalOrderStruct, err error) {

	var GlobalPacketDEC GlobalOrderStruct
	err = json.Unmarshal(JsonPacket, &GlobalPacketDEC)
	return GlobalPacketDEC, err
}
