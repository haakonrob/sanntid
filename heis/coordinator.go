package main

import(
"fmt"
"encoding/json"
//"./fsm"
heis "./heisdriver"
_"os"
"time"
)

const 

const ( 
	MAX_NUM_ELEVS = 10
 	N_FLOORS = heis.N_FLOORS
	UP = heis.BUTTON_CALL_UP
	DOWN = heis.BUTTON_CALL_DOWN
	COMMAND = heis.BUTTON_COMMAND
)

// INCLUDED IN FSM.go!
type LocalOrderStruct struct {
	Pending [3][NUM_FLOORS] bool
	Completed [3][NUM_FLOORS] bool
	prevFloor int
	direction heis.ElevMotorDirection
}


type GlobalOrderStruct struct {
	Available [2][NUM_FLOORS] bool
	Taken [2][NUM_FLOORS] int
	Timestamps [2][NUM_FLOORS] uint
	Clock uint
	Scores [MAX_NUM_ELEVS][2][NUM_FLOORS] int
	LocalOrdersBackup[MAX_NUM_ELEVS] LocalOrderStruct
}

//var LocalOrders fsm.LocalOrderStruct
var LocalOrders LocalOrderStruct
var GlobalOrders GlobalOrderStruct
var online bool
var localElevIndex int
var activeElevs []int

/******REMOVE*********/
var prevFloor int
var currDir heis.ElevMotorDirection
 /*****REMOVE*********/

/*********************************
Testing for network encoding
*********************************/
var str string

/*********************************
Main 
*********************************/

func main(){
	
	
	localElevIndex = 0;
	activeElevs = make([]int, 0, MAX_NUM_ELEVS)
	
	
	// FSM NEEDS initElev() so that it's at a floor

	// Sets all Taken values to -1. Not o because 0 can be a elevIndex
	for _, ordertype := range([]int{UP, DOWN}) {
    	for _, floor := range(GlobalOrders.Taken[ordertype]) {
    		GlobalOrders.Taken[ordertype][floor] = -1
		}
	}
	
	orderChan := make(chan heis.Order, 5) 
	eventChan := make(chan heis.Event, 5)
	/*****ADD*****
	fsmChan := make(chan fsm.FsmPackageStruct)
	networkChan := make(chan string)
	**************/
	
	
	
	//nextOrder := getNextOrder()
	//go doOrder(nextOrder)
	//if online {
	//	recvOrdersFromNetwork()
	//}
	//scoreOrders()
	//if online {
	//	sendOrdersToNetwork()
	//}
	
	
	go heis.Poller(orderChan, eventChan)
	/*****ADD*****
	go networkMonitor(network)
	go fsm.Fsm(eventChan)
	**************/
	timestamp := time.Now()
	for {
		// online, localElevIndex,  = networkStatus()
		select {
			case order := <-orderChan:
				fmt.Println(order)
				//addNewOrder(order)
			case ev := <-eventChan:
				fmt.Println(ev)
			/*****ADD*****
			case msg := <-networkChan:
				mergeOrders() 
			go fsm.Fsm(eventChan)
			**************/
			default:
				getNextOrder()
				if time.Since(timestamp) > time.Second*100 {
					// testing getNextOrder(). Works!
					fmt.Println(GlobalOrders.Available)
					fmt.Println(GlobalOrders.Taken)
					fmt.Println(getNextOrder())
					timestamp = time.Now()
				}
				// set nextOrder?
				continue
		}
	}
}

/*********************************
Functions
*********************************/

func addNewOrder(order heis.Order){
	/*
	if order.OrderType == 
	Try to add to global orders


	*/
	LocalOrders.Pending[order.OrderType][order.Floor] = true
	LocalOrders.Completed[order.OrderType][order.Floor] = false
}

// getNextOrder() will be replaced by updateLocalOrders()
func getNextOrder() (int, int, bool) {
	//IMPROVEMENTS:
	//Collect and choose best option instead of taking first one

	// If elev score is best in table AND not 0, it claims the order
	for _, ordertype := range([]int{UP, DOWN}){
		for floor := 0; floor<NUM_FLOORS; floor++ {
			if GlobalOrders.Available[ordertype][floor] {
					if isBestScore(ordertype, floor) {
						GlobalOrders.Available[ordertype][floor] = false
						GlobalOrders.Taken[ordertype][floor] = localElevIndex
						GlobalOrders.Timestamps[ordertype][floor] = GlobalOrders.Clock
						return ordertype, floor, true
					}
				}
		}
	}
	return -1,-1, false
}

func scoreOrders(){
	for _, ordertype := range([]int{UP, DOWN}){
		for floor := 0; floor<NUM_FLOORS; floor++ {
			//INSERT COST FUNC HERE
			GlobalOrders.Scores[localElevIndex][ordertype][floor] = 10;
		}
	}
}

func isBestScore(ordertype int, floor int) bool{
	// returns false if it finds a better competitor, else returns true. 
	// If this is somehow called when the network is down, ignore globalorders.
	if len(activeElevs) == 0 || !online {
		return true
	}
	for _, extElevIndex := range(activeElevs) {
		if GlobalOrders.Scores[localElevIndex][ordertype][floor] < GlobalOrders.Scores[extElevIndex][ordertype][floor]{
			return false
		}
	}
	return true
}



func completeOrder(order heis.Order){
	LocalOrders.Pending[order.OrderType][order.Floor] = false
	LocalOrders.Completed[order.OrderType][order.Floor] = true
	// Turn off light?
	// GlobalOrders.Available[i] = false
	// GlobalOrders.Taken[i] = -1
	// stop and open doors
}

func sendOrdersToNetwork() bool {
	// could be replaced with an encode/decode function
	/*
	msg, err := json.Marshal(GlobalOrders)
	if err == nil {
		network<- msg
		// wait for ACK
	}
	*/

	msg, _ := json.Marshal(GlobalOrders)
	str = string(msg)
	return true
}

func recvOrdersFromNetwork(network chan []byte) bool {
	// could be replaced with an encode/decode function
	/*
	temp := make(GlobalOrderStruct)
	select {
	case msg := <-network:
		err = json.Unmarshal([]byte(str), &temp)
		if err == nil {
			Orders = mergeOrders(temp)
			return true
		}
	default:
		return false
	}
	*/
	 _ = json.Unmarshal([]byte(str), &GlobalOrders)
	fmt.Println(GlobalOrders)
	return true
}

func mergeOrders(newGlobalOrders GlobalOrderStruct){
	GlobalOrders = newGlobalOrders
	//^uint(0) gives the maximum value for uint
	GlobalOrders.Clock = (GlobalOrders.Clock+1)%(^uint(0))
	// If an order is listed as taken, but this elev has completed it, the order is removed from globalorders
	for _, ordertype := range([]int{UP, DOWN}){
		for floor := 0; floor<NUM_FLOORS; floor++ {
			if newGlobalOrders.Taken[ordertype][floor] >= 0 && LocalOrders.Completed[ordertype][floor] {
				GlobalOrders.Taken[ordertype][floor] = -1
				LocalOrders.Completed[ordertype][floor] = false
			}
		}
	}
}
