package main

import (
	"./fsm"
	heis "./heisdriver" //"./simulator/client"
	"./network"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	//"os/exec"
)

/******************************************
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
	Available  	[2][N_FLOORS]bool      //'json:"Available"'
	Taken      	[2][N_FLOORS]bool      //'json:"Taken"'
	Timestamps 	[2][N_FLOORS]time.Time //'json:"Timestamps"'
	Scores            map[string][2][N_FLOORS]int        //[MAX_NUM_ELEVS][2][N_FLOORS]int    //'json:"Scores"'
	//LocalOrdersBackup [MAX_NUM_ELEVS]fsm.LocalOrderState //'json:"LocalOrdersBackup"'
	
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
	GlobalOrders.Scores = make(map[string][2][N_FLOORS]int)

	orderChan := make(chan heis.Order, 5)
	completedOrderChan := make(chan fsm.LocalOrderState)
	eventChan := make(chan heis.Event, 5)
	fsmChan := make(chan fsm.LocalOrderState)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	networkCh := make(chan string)

	heis.ElevInit()
	go heis.Poller(orderChan, eventChan)
	go fsm.Fsm(eventChan, fsmChan, completedOrderChan)
	go network.Monitor(networkCh, incomingCh, outgoingCh)

	var changesMade bool
	for {
		if (changesMade) {
			changesMade = false
			orderIterate(COMMAND, N_FLOORS, setLights)
			// backup
			if online {
				for i:=0;i<1;i++ {
					outgoingCh <- EncodeGlobalPacket()
				}
			}
		}
		
		select {
		case status := <-networkCh:
			online, localID, activeElevs = decodeNetworkStatus(status)
			fmt.Println(activeElevs)
			fmt.Println("Online: ", online)
		case msg := <-incomingCh:
			//fmt.Println("Network pkt ")
			var GlobalPacketDEC GlobalOrderStruct
			err := json.Unmarshal(msg, &GlobalPacketDEC)
			if err == nil {	
				GlobalOrders = GlobalPacketDEC
				//fmt.Println("Good network package", err, GlobalPacketDEC)
			} else {
				fmt.Println("Bad network package", err)
				break
			}
			
			orderIterate(DOWN, N_FLOORS, mergeOrders)
			
			switch (GlobalOrders.SenderId) {
			case localID:
				if orderIterate(DOWN, N_FLOORS,takeGlobalOrder) {
					changesMade = true
					//updateLocalPendingOrders()
					GlobalOrders.SenderId = localID
					fsmChan <- LocalOrders

					fmt.Println("Taken new order: ", GlobalOrders.Taken)
					fmt.Println("Local Taken new order: ", LocalOrders.Pending)
				} else {
					fmt.Println("None to take: ", GlobalOrders.Taken)
				}
			default:
				if orderIterate(DOWN, N_FLOORS, globalOrdersAvailable){
					changesMade = true 
					orderIterate(DOWN, N_FLOORS, scoreAvailableOrder)
					//fmt.Println("Scored")
				} else {
					//fmt.Println("No Orders available")
					fmt.Println(GlobalOrders.Taken)
				}
			}
	
		case newOrder := <-orderChan:
			switch (online) {
			case true:
				if newOrder.OrderType != COMMAND{
					fmt.Println("New global order.")
					changesMade = addNewGlobalOrder(newOrder)
					fmt.Println("Changed: ", changesMade)
					GlobalOrders.SenderId = localID
				} else {
					fmt.Println("New local order.")
					changesMade = addNewLocalOrder(newOrder)
					fmt.Println("Changed: ", changesMade)
					fsmChan <- LocalOrders
				}
				// fsm will be updated when packet comes around again
			case false:
				changesMade = addNewLocalOrder(newOrder)
				fmt.Println("Changed: ", changesMade)
				fsmChan <- LocalOrders
			}
			changesMade = true

			
		case newLocalOrders := <-completedOrderChan:
			LocalOrders.Completed = newLocalOrders.Completed
			switch (online) {
			case true:
				changesMade = orderIterate(COMMAND, N_FLOORS, completeGlobalOrders)
				GlobalOrders.SenderId = localID

			case false:
				changesMade = orderIterate(COMMAND, N_FLOORS, completeLocalOrders)
				fsmChan <- LocalOrders
			}

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

func addNewLocalOrder(order heis.Order)(bool) {
	ordertype := order.OrderType
	floor := order.Floor
	stamp := time.Now()
	result := false
	switch ordertype {
	case UP, DOWN, COMMAND:
		if !LocalOrders.Pending[ordertype][floor] {
			LocalOrders.Pending[ordertype][floor] = true
			LocalOrders.Completed[ordertype][floor] = false
			LocalOrders.Timestamps[ordertype][floor] = stamp
			result = true
		}

	default:
		fmt.Println("Invalid OrderType in addNewLocalOrder()")
	}
	return result

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
				scoreAvailableOrder(order.OrderType, order.Floor)	
				fmt.Println("My score: ", order, GlobalOrders.Scores[localID][order.OrderType][order.Floor])			
			} else {
				temp[ordertype][floor] = 0
				GlobalOrders.Scores[ID] = temp
			}
			//fmt.Println("My score: ", order, GlobalOrders.Scores[localID][order.OrderType][order.Floor])
		}
	
	default:
		fmt.Println("Invalid OrderType in addnewglobalorder()")
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
		LocalOrders.Completed[ordertype][floor] = false
		return true
	}
	return false
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
func setLights(ordertype heis.ElevButtonType, floor int)(bool) {
	var val bool
	switch ordertype {
	case COMMAND:
		val = LocalOrders.Pending[ordertype][floor]
	case UP, DOWN:
		if online {
			val = GlobalOrders.Available[ordertype][floor] || GlobalOrders.Taken[ordertype][floor]
		} else {
			val = LocalOrders.Pending[ordertype][floor]
		}
		heis.ElevSetButtonLamp(ordertype, floor, val)
	default:
		fmt.Println("Invalid ordertype in setLights()")
	}
	heis.ElevSetButtonLamp(ordertype, floor, val)
	return true

}
//2 x N_FLOORS
func scoreAvailableOrder(ordertype heis.ElevButtonType, floor int)(bool) {
	if scores, ok := GlobalOrders.Scores[localID]; ok {
		//fmt.Println("Actually scored")
		pFloor, dir := LocalOrders.PrevFloor, LocalOrders.Direction
		floorDiff := (floor - pFloor)	
		if floorDiff < 0 {floorDiff = -floorDiff}
		randNum := int(byte(localID[len(localID)-1]))
		scores[ordertype][floor] = 200-floorDiff + (floor - pFloor)*int(dir)*10 - randNum //COST FUNC
		GlobalOrders.Scores[localID] = scores
		//fmt.Println(randNum)
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
			} else if GlobalOrders.Scores[localID][ordertype][floor] == 0 {
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
		LocalOrders.Pending[ordertype][floor] = GlobalOrders.Taken[ordertype][floor]
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
