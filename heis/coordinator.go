package main

import(
"fmt"
"encoding/json"
heis "./heisdriver"
_"os"
"time"
)

// Number of elevators in system
const MAX_NUM_ELEVS = 10

type orderIndex int
const ( 
	ORDER_UP1 orderIndex = iota
    ORDER_UP2
    ORDER_UP3
    ORDER_DOWN2
    ORDER_DOWN3
    ORDER_DOWN4
    ORDER_COMM1
    ORDER_COMM2
    ORDER_COMM3
    ORDER_COMM4
    ORDER_MAXINDEX
)

type OrderStruct struct {
	//available orders
	Available [ORDER_MAXINDEX] bool
	//taken orders
	Taken [ORDER_MAXINDEX] int
	//order timestamps
	Timestamps [ORDER_MAXINDEX] uint
	//clock - number of times the package has been passed on
	Clock uint
	//elevator scores
	Scores [MAX_NUM_ELEVS][ORDER_MAXINDEX] int
	// In case of a crash, all elevators know about all local commands
	//AllCommands [MAX_NUM_ELEVS][ORDER_MAXINDEX] bool
}

var Orders OrderStruct
var online bool
var localElevIndex int
var activeElevs []int
var prevFloor int
var currDir heis.ElevMotorDirection
/*********************************
Testing for network encoding
*********************************/
var str string

/*********************************
Main 
*********************************/

func main(){
	
	// localElevIndex needs to be set by the network function
	localElevIndex = 0;
	activeElevs = make([]int, 0, MAX_NUM_ELEVS)
	
	heis.ElevInit()
	// initElev() so that it's at a floor
	for i, _ := range(Orders.Taken) {
    	Orders.Taken[i] = -1
	}
	
	
	orderChan := make(chan heis.Order, 5) 
	eventChan := make(chan heis.Event, 5)
	//network := make(chan string)
	
	
	//nextOrder := getNextOrder()
	//go doOrder(nextOrder)
	//if online {
	//	recvOrdersFromNetwork()
	//}
	//scoreOrders()
	//if online {
	//	sendOrdersToNetwork()
	//}
	
	
	// go networkMonitor(network)
	go heis.Poller(orderChan, eventChan)
	
	timestamp := time.Now()
	for {
		// networkUpdate, returns local elev index, active elevs
		// 
	
		
		select {
			case order := <-orderChan:
				fmt.Println(order)
				addNewOrder(order)
			case ev := <-eventChan:
				fmt.Println(ev)
			default:
				if time.Since(timestamp) > time.Second*100 {
					// testing getNextOrder(). Works!
					fmt.Println(Orders.Available)
					fmt.Println(Orders.Taken)
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

func getNextOrder() (orderIndex, bool) {
	//IMPROVEMENTS:
	//Collect and choose best option instead of taking first one

	// If elev score is best in table AND not 0, it claims the order
	for i:=ORDER_UP1; i<ORDER_MAXINDEX; i++ {
		if Orders.Available[i] {
				if isBestScore(i) {
					Orders.Available[i] = false
					Orders.Taken[i] = localElevIndex
					Orders.Timestamps[i] = Orders.Clock
					return i, true
				}
			}
	}
	return ORDER_MAXINDEX, false
}

func isBestScore(i orderIndex) bool{
	// returns false if it finds a better competitor, else returns true.
	if len(activeElevs) == 0 {
		return true
	}
	for _, extElevIndex := range(activeElevs) {
		if Orders.Scores[localElevIndex][i] < Orders.Scores[extElevIndex][i]{
			return false
		}
	}
	return true
}

func scoreOrders(){
	for i:=ORDER_UP1; i<ORDER_MAXINDEX; i++ {
		Orders.Scores[localElevIndex][i] = 10
	}
}

func addNewOrder(order heis.Order){
	var i int
	switch(order.OrderType){
	case 0:
		i = order.Floor + int(ORDER_UP1)
	case 1:
		i = order.Floor + int(ORDER_DOWN2)-1
	case 2:
		i = order.Floor + int(ORDER_COMM1)
	}
	if Orders.Taken[i] < 0 {
		Orders.Available[i] = true
	}
}

func completeOrder(i orderIndex){
	Orders.Available[i] = false
	Orders.Taken[i] = -1
	// stop and open doors
}

func sendOrdersToNetwork() bool {
	/*
	msg, err := json.Marshal(Orders)
	if err == nil {
		network<- msg
		// wait for ACK
		
	}
	*/

	msg, _ := json.Marshal(Orders)
	str = string(msg)
	return true
}

func recvOrdersFromNetwork(network chan []byte) bool {
	/*
	temp := make(OrderStruct)
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
	 _ = json.Unmarshal([]byte(str), &Orders)
	fmt.Println(Orders)
	return true
}

func mergeOrders(newOrders OrderStruct){
	// This situation suggests that the elevator has completed its order and needs to update the others
	for i:=ORDER_UP1; i<ORDER_MAXINDEX; i++ {
		if !(newOrders.Taken[i] == localElevIndex && Orders.Taken[i] == -1) {
			Orders.Taken[i] = newOrders.Taken[i]
		}
	}
	Orders.Available = newOrders.Available		
	Orders.Timestamps = newOrders.Timestamps
	//^uint(0) gives the maximum value for uint
	Orders.Clock = (newOrders.Clock+1)%(^uint(0))
	Orders.Scores = newOrders.Scores
}
