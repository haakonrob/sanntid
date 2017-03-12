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
	SenderId          string
	Available  [2][N_FLOORS]bool      //'json:"Available"'
	Taken      [2][N_FLOORS]bool      //'json:"Taken"'
	Timestamps [2][N_FLOORS]time.Time //'json:"Timestamps"'
	Scores            map[string][2][N_FLOORS]int        //[MAX_NUM_ELEVS][2][N_FLOORS]int    //'json:"Scores"'
	LocalOrdersBackup [MAX_NUM_ELEVS]fsm.LocalOrderState //'json:"LocalOrdersBackup"'
	
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

	activeElevs = make([]string, 0, MAX_NUM_ELEVS)

	orderChan := make(chan heis.Order, 5)
	eventChan := make(chan heis.Event, 5)
	fsmChan := make(chan fsm.LocalOrderState)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	networkCh := make(chan string)

	heis.ElevInit()
	go heis.Poller(orderChan, eventChan)
	go fsm.Fsm(eventChan, fsmChan)
	go network.Monitor(networkCh, incomingCh, outgoingCh)

	var changesMade bool
	for {
		if (changesMade) && online {
			changesMade = false
			for i:=0;i<5;i++ {
				GlobalOrders.SenderId = localID
				outgoingCh <- EncodeGlobalPacket()
			}
		}
		
		select {
		case status := <-networkCh:
			online, localID, activeElevs = decodeNetworkStatus(status)
		
		case msg := <-incomingCh:
			//oldGlobalOrders := GlobalOrders
			GlobalOrders, err := DecodeGlobalPacket(msg)
			if err != nil {
				fmt.Println("Bad network package")
				break
			}
			
			orderIterate(DOWN, N_FLOORS, mergeOrders)
			
			switch (GlobalOrders.SenderId) {
			case localID:
				if orderIterate(DOWN, N_FLOORS,takeGlobalOrder) {
					changesMade = true
					updateLocalPendingOrders()
				}
			default:
				if orderIterate(DOWN, N_FLOORS, globalOrdersAvailable){
					changesMade = true 
					orderIterate(DOWN, N_FLOORS, scoreAvailableOrders)
				}
			}
	
		case newOrder := <-orderChan:
			switch (online) {
			case true && newOrder.OrderType != COMMAND:
				changesMade = addNewGlobalOrder(newOrder)
				// fsm will be updated when packet comes around again
			case false || newOrder.OrderType == COMMAND:
				_ = addNewLocalOrder(newOrder)
				fsmChan <- LocalOrders
			}
			
		case newLocalOrders := <-fsmChan:
			LocalOrders.Completed = newLocalOrders.Completed
			switch (online) {
			case true:
				changesMade = orderIterate(DOWN, N_FLOORS, completeGlobalOrders)

			case false:
				orderIterate(COMMAND, N_FLOORS, completeLocalOrders)
				fsmChan <- LocalOrders
			}

		default:

			//timeout error handling
			//if global taken too long
			//move from taken to avaliable
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

func addNewLocalOrder(order heis.Order)(bool) {
	ordertype := order.OrderType
	floor := order.Floor
	stamp := time.Now()
	switch ordertype {
	case UP, DOWN, COMMAND:
		LocalOrders.Pending[ordertype][floor] = true
		LocalOrders.Completed[ordertype][floor] = false
		LocalOrders.Timestamps[ordertype][floor] = stamp

	default:
		fmt.Println("Invalid OrderType in addNewLocalOrder()")
	}
	return true

}

func updateLocalPendingOrders(){
	LocalOrders.Pending[UP] = GlobalOrders.Taken[UP]
	LocalOrders.Pending[DOWN] = GlobalOrders.Taken[DOWN]
}

func addNewGlobalOrder(order heis.Order)(bool) {
	ordertype := order.OrderType
	floor := order.Floor
	stamp := time.Now()

	switch ordertype {
	case UP, DOWN :
		GlobalOrders.Available[ordertype][floor] = true
		GlobalOrders.Taken[ordertype][floor] = false
		GlobalOrders.Timestamps[ordertype][floor] = stamp
		
		for _, ID := range(activeElevs) {
			temp := GlobalOrders.Scores[ID]
			if ID == localID {
				temp[ordertype][floor] = 10 //add score func
			} else {
				temp[ordertype][floor] = 0
			}
			GlobalOrders.Scores[ID] = temp
		}

	default:
		fmt.Println("Invalid OrderType in updateglobalorders()")
	}
	return true

}


/****************************************************
 Functions for handling LocalOrders and GLobalOrders
*****************************************************/

type orderLogic func(heis.ElevButtonType, int)(bool)

//Iterates over specified intevals with a function, 
//returns true if a single iteration resulted in true
func orderIterate( oEnd heis.ElevButtonType, fEnd int, orderFun orderLogic )(bool){
	result := false
	for o := UP; o <= oEnd; o++ {
		for f := 0; f < fEnd; f++ {
			if orderFun(o,f){
				result = true
			}
		}
	}
	return result		
} 

func globalOrdersAvailable(ordertype heis.ElevButtonType, floor int)(bool){
	if GlobalOrders.Available[ordertype][floor]{
		return true
	}
	return false
}

//3 x N_FLOORS
func completeLocalOrders(ordertype heis.ElevButtonType, floor int)(bool) {
	if LocalOrders.Completed[ordertype][floor]{
		LocalOrders.Pending[ordertype][floor] = false
		return true
	}
	return false
	
	/*
	for o := UP; o <= COMMAND; o++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			LocalOrders.Completed[ordertype][floor] = newLocalOrders.Completed[ordertype][floor]
			LocalOrders.Pending[o][floor] = LocalOrders.Pending[o][floor] && !(completed[o][floor])
		}
	}
*/
}


// 3 x N_FLOORS
func completeGlobalOrders(o heis.ElevButtonType, floor int)(bool) {
	completed := LocalOrders.Completed

	if completed[o][floor]{
		switch o {
		case UP, DOWN:
			GlobalOrders.Available[o][floor] = false
			GlobalOrders.Taken[o][floor] = false
			LocalOrders.Pending[o][floor] = false
			LocalOrders.Completed[o][floor] = false
			return true
		
		case COMMAND :
			LocalOrders.Pending[o][floor] = false
			LocalOrders.Completed[o][floor] = false
			return true

		default:
			fmt.Println("Bad ordertype in completeGlobalOrders()")
		}	
	}
	return false
	
}

/*************************************************/

//3 x N_FLOORS
func setLight(ordertype heis.ElevButtonType, floor int)(bool) {
	switch ordertype {
	case COMMAND:
		val := LocalOrders.Pending[ordertype][floor]
		heis.ElevSetButtonLamp(ordertype, floor, val)
	case UP, DOWN:
		val := GlobalOrders.Available[ordertype][floor] || GlobalOrders.Taken[ordertype][floor]
		heis.ElevSetButtonLamp(ordertype, floor, val)
	}
	return true

}
//2 x N_FLOORS
func scoreAvailableOrders(ordertype heis.ElevButtonType, floor int)(bool) {
	if scores, ok := GlobalOrders.Scores[localID]; ok {
		scores[ordertype][floor] = 10 ////INSERT COST FUNC HERE
		return true
	} else {
		return false
	}
}

// 2 x N_FLOORS
func isBestScore(ordertype heis.ElevButtonType, floor int) (bool) {
	// returns false if it finds a better competitor, else returns true.

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

//2 x N_FLOORS
func takeGlobalOrder(ordertype heis.ElevButtonType, floor int)(bool) {
	changesMade := false
	if isBestScore(ordertype, floor) && GlobalOrders.Available[ordertype][floor] {
		GlobalOrders.Available[ordertype][floor] = false
		GlobalOrders.Taken[ordertype][floor] = true
		GlobalOrders.Timestamps[ordertype][floor] = time.Now()
		for _, elev := range(activeElevs) {
			temp := GlobalOrders.Scores[elev]
			temp[ordertype][floor] = 0
			GlobalOrders.Scores[elev] = temp
		}
		changesMade = true
	}
	return changesMade
}

//2 x N_FLOORS
func mergeOrders(ordertype heis.ElevButtonType, floor int)(bool) {
	// If an order is listed as taken, but this elev has completed it, the order is removed from globalorders
	if GlobalOrders.Taken[ordertype][floor] && LocalOrders.Completed[ordertype][floor] {
		GlobalOrders.Taken[ordertype][floor] = false
		LocalOrders.Completed[ordertype][floor] = false
		return true
	}
	return false
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
