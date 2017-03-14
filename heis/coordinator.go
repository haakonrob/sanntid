package main

import (
	"./fsm"
	heis "./heisdriver" //"./simulator/client"
	"./network"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"os"
	"os/exec"
	"io/ioutil"
)

/******************************************
TODO - Fault tolerance
Add timestamping logic
Data logging for backup in case of crash/termination


DISCUSS
Timestamps in Taken instead of heisnr?
******************************************/

const (
	MAX_NUM_ELEVS = 10
	N_FLOORS      = heis.N_FLOORS
	UP            = heis.BUTTON_CALL_UP
	DOWN          = heis.BUTTON_CALL_DOWN
	COMMAND       = heis.BUTTON_COMMAND
	loopBack	  = true
	subnet		  = "localhost"
)

type GlobalOrderStruct struct {
	SenderId          string
	Available  	[2][N_FLOORS]bool      //'json:"Available"'
	Taken      	[2][N_FLOORS]bool      //'json:"Taken"'
	Timestamps 	[2][N_FLOORS]time.Time //'json:"Timestamps"'
	Scores            map[string][2][N_FLOORS]int      //'json:"Scores"'
	
} //

var (
	GlobalOrders GlobalOrderStruct
	LocalOrders fsm.LocalOrderState
	online      bool
	localID     string
	activeElevs []string
	changesMade bool
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

	networkCh := make(chan string)
	//bcastEN := make(chan bool)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	timeoutChan	:= make(chan string)

	heis.ElevInit()
	go fsm.Fsm(eventChan, fsmChan, completedOrderChan)
	go network.Monitor(networkCh, true, "sanntidsal", incomingCh, outgoingCh)
	go heis.Poller(orderChan, eventChan)
	go timeOut(timeoutChan)

	printOrders()


	if len(os.Args)>1 {
		loadFromBackup(fsmChan)
	} 
	
	for {
		if (changesMade) {

			changesMade = false
			orderIterate(COMMAND, N_FLOORS, setLights)
			data, _ := json.Marshal(LocalOrders)
			_ = ioutil.WriteFile("./backupdata", data, 0644)
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
			//temp := GlobalOrders.Scores
			for id, _ := range(GlobalOrders.Scores) {
				_, idExists := GlobalOrders.Scores[id]
				if !idExists {
					GlobalOrders.Scores[id] = [2][N_FLOORS]int{}
				}
			}
			//fmt.Println(GlobalOrders.Scores[localID])

		case msg := <-incomingCh:
			var GlobalPacketDEC GlobalOrderStruct
			err := json.Unmarshal(msg, &GlobalPacketDEC)
			if err == nil {	
				GlobalOrders = GlobalPacketDEC
				//fmt.Println("New Orders: ", GlobalOrders.Scores)
				printOrders()
			} else {
				fmt.Println("Bad network package", string(msg))
				break
			}
			
			orderIterate(DOWN, N_FLOORS, mergeOrders)
			GlobalOrders.Available = GlobalPacketDEC.Available
			
			switch (GlobalOrders.SenderId) {
			case localID:
				if orderIterate(DOWN, N_FLOORS, takeGlobalOrder) {
					changesMade = true
					GlobalOrders.SenderId = localID
					fsmChan <- LocalOrders

				} else {
					//fmt.Println("None to take: ", GlobalOrders.Taken)
				}
			default:
					changesMade = orderIterate(DOWN, N_FLOORS, scoreAvailableOrder)
			}
	
		case newOrder := <-orderChan:
			switch (online) {
			case true:
				if newOrder.OrderType != COMMAND{
					changesMade = addNewGlobalOrder(newOrder)
					GlobalOrders.SenderId = localID
				} else {
					changesMade = addNewLocalOrder(newOrder)
					fsmChan <- LocalOrders
				}
				// fsm will be updated when packet comes around again
			case false:
				changesMade = addNewLocalOrder(newOrder)
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


		case timeOut:= <-timeoutChan:
			switch (timeOut){
			case "LOCAL_TIMEOUT":
				BackupRestart()
			case "GLOBAL_TIMEOUT":
				changesMade = true
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

func addNewGlobalOrder(order heis.Order)(bool) {
	ordertype := order.OrderType
	floor := order.Floor
	//stamp := time.Now()
	switch ordertype {
	case UP, DOWN :
		if !GlobalOrders.Taken[ordertype][floor] {
			GlobalOrders.Available[ordertype][floor] = true
			
			for _, ID := range(activeElevs) {
				scores, _ := GlobalOrders.Scores[ID]
				if ID == localID {
					scoreAvailableOrder(order.OrderType, order.Floor)			
				} else {
					scores[ordertype][floor] = 0
					GlobalOrders.Scores[ID] = scores
				}
				
			}
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
		LocalOrders.Timestamps[ordertype][floor] = time.Time{}
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
			GlobalOrders.Timestamps[o][floor] = time.Time{}
			LocalOrders.Pending[o][floor] = false
			LocalOrders.Completed[o][floor] = false
			LocalOrders.Timestamps[o][floor] = time.Time{}

			return true
		
		case COMMAND :
			LocalOrders.Pending[o][floor] = false
			LocalOrders.Completed[o][floor] = false
			LocalOrders.Timestamps[o][floor] = time.Time{}

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
			val = GlobalOrders.Taken[ordertype][floor]
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
		if GlobalOrders.Available[ordertype][floor] {
			//fmt.Println("Actually scored")
			pFloor, dir := LocalOrders.PrevFloor, LocalOrders.Direction
			floorDiff := (floor - pFloor)	
			if floorDiff < 0 {floorDiff = -floorDiff}
			//randNum := int(byte(localID[len(localID)-1]))
			scores[ordertype][floor] = 200-floorDiff + (floor - pFloor)*int(dir)*10 //COST FUNC
			GlobalOrders.Scores[localID] = scores
			//fmt.Println("Available score ", scores[ordertype][floor], scores)
			return true
		}
	}
	return false
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
		//fmt.Println("Taking an order")
		GlobalOrders.Available[ordertype][floor] = false
		GlobalOrders.Taken[ordertype][floor] = true
		GlobalOrders.Timestamps[ordertype][floor] = time.Now()
		LocalOrders.Pending[ordertype][floor] = GlobalOrders.Taken[ordertype][floor]
		LocalOrders.Timestamps[ordertype][floor] = GlobalOrders.Timestamps[ordertype][floor]
		for _, elev := range(activeElevs) {
			temp, idExists := GlobalOrders.Scores[elev]
			if !idExists {

			}
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


func isLocalTimeout(ordertype heis.ElevButtonType, floor int)bool{
	if LocalOrders.Pending[ordertype][floor] {
		if time.Since(LocalOrders.Timestamps[ordertype][floor]) > time.Second*20 {
			if online {
				LocalOrders.Pending[ordertype][floor] = false			
			}
			LocalOrders.Timestamps[ordertype][floor] = time.Now()
			data, _ := json.Marshal(LocalOrders)
			_ = ioutil.WriteFile("./backupdata", data, 0644)
			return true
		}
	}
	return false
}

func isGlobalTimeout(ordertype heis.ElevButtonType, floor int)bool{
	copyOrder := heis.Order{ordertype, floor}
	if GlobalOrders.Taken[ordertype][floor] {
		if (time.Since(GlobalOrders.Timestamps[ordertype][floor]) > time.Second*10) {
			switch online {
			case true:
				GlobalOrders.Taken[ordertype][floor] = false
				GlobalOrders.Timestamps[ordertype][floor] = time.Now()

				copyOrder.Floor = floor
				copyOrder.OrderType = ordertype

				addNewGlobalOrder(copyOrder)
			case false:
				//dontcare

			}
			return true 
		}
	}
	return false
}





func printOrders(){

	fmt.Println("LocalID: 	", localID)
	fmt.Println("GlobalOrders Available:")
	for i := 0; i < 2; i++{
		fmt.Println("                       		", GlobalOrders.Available[i])
	}
	fmt.Println("GlobalOrders Taken:")
	for i := 0; i < 2; i++{
		fmt.Println("                       		", GlobalOrders.Taken[i])
	}
	fmt.Println("Scores:")
	//fmt.Println("                       		", GlobalOrders.Scores)
	fmt.Println("           	           		", GlobalOrders.Scores[localID])
	
	for _, elevID := range activeElevs {
		fmt.Println("                       		", GlobalOrders.Scores[elevID])
	}


	fmt.Println("LocalOrders Pending:")
	for i := 0; i < 3; i++{
		fmt.Println("                       		", LocalOrders.Pending[i])
	}

}


func timeOut(timeoutChan chan string){
	for{
		time.Sleep(time.Second)
		if orderIterate(COMMAND, N_FLOORS, isLocalTimeout) {
			timeoutChan<-"LOCAL_TIMEOUT" 
		} else if orderIterate(DOWN, N_FLOORS, isGlobalTimeout){
			timeoutChan<-"GLOBAL_TIMEOUT"
		}
	}
}

func BackupRestart(){
	//bcastEN<- false
	fmt.Println("Local timeout, unknown error.")

	backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run coordinator.go ./backupdata")
	backup.Run()
	for o := UP; o <= COMMAND; o++ {
		for f := 0; f < N_FLOORS; f++ {
			LocalOrders.Timestamps[o][f] = time.Now()
		}
	}
	//LocalOrders.Timestamps[ordertype][floor] = time.Time{}
	data, _ := json.Marshal(LocalOrders)
	_ = ioutil.WriteFile("./backupdata", data, 0644)
	os.Exit(0)

}

func loadFromBackup(fsmChan chan fsm.LocalOrderState){
		backupFilePath := os.Args[1]
		data, _ := ioutil.ReadFile(backupFilePath)
		temp := LocalOrders
		err := json.Unmarshal(data, &temp)
		if (err == nil){
			LocalOrders = temp
			changesMade = true
			fsmChan<- LocalOrders
		}
}
