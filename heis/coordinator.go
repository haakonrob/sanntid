package main

import (
	"./fsm"
	heis "./heisdriver" //"./simulator/client"
	"./network"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
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
	loopBack      = false
	subnet        = "sanntidsal"
)

type GlobalOrderStruct struct {
	SenderId   string
	Available  [2][N_FLOORS]bool           //'json:"Available"'
	Taken      [2][N_FLOORS]bool           //'json:"Taken"'
	Timestamps [2][N_FLOORS]time.Time      //'json:"Timestamps"'
	Scores     map[string][2][N_FLOORS]int //'json:"Scores"'

} //

var (
	GlobalOrders GlobalOrderStruct
	LocalOrders  fsm.LocalOrderState
	online       bool
	localID      string
	activeElevs  []string
	changesMade  bool
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

	networkCh := make(chan network.Status)
	//bcastEN := make(chan bool)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	timeoutChan := make(chan string)

	heis.ElevInit()
	go fsm.Fsm(eventChan, fsmChan, completedOrderChan)
	go network.Monitor(networkCh, loopBack, "sanntidsal", incomingCh, outgoingCh)
	go heis.Poller(orderChan, eventChan)
	go timeOut(timeoutChan)

	printOrders()

	if len(os.Args) > 1 {
		loadFromBackup(fsmChan)
		//fmt.Println("Loaded From backup", LocalOrders)
	}

	for {
		if changesMade {
			//printOrders()
			//fmt.Println("CHANGE MADE")
			changesMade = false
			fmt.Println("IM GOING THROUGH CHANGES")
			orderIterate(COMMAND, N_FLOORS, setLights)

			data, _ := json.Marshal(LocalOrders)
			_ = ioutil.WriteFile("./backupdata", data, 0644)
			if online {
				for i := 0; i < 1; i++ {
					outgoingCh <- EncodeGlobalPacket()
				}
			}
		}

		select {
		case status := <-networkCh:
			online, localID, activeElevs = status.Online, status.LocalID, status.ActiveIDs
			fmt.Println("Online: ", online)

		case msg := <-incomingCh:
			var GlobalPacketDEC GlobalOrderStruct
			err := json.Unmarshal(msg, &GlobalPacketDEC)
			if err == nil {
				printOrders()
				mergeGlobalOrders(GlobalPacketDEC)
				orderIterate(DOWN, N_FLOORS, scoreAvailableOrder)

				//GlobalOrders = GlobalPacketDEC
				//fmt.Println("incoming: ")
				//printOrders()
			} else {
				fmt.Println("Bad network package", string(msg))
				break
			}
			orderIterate(DOWN, N_FLOORS, mergeWithLocOrders)

			switch GlobalOrders.SenderId {
			case localID:
				if orderIterate(DOWN, N_FLOORS, takeGlobalOrder) {
					fmt.Println("TAKEN")
					printOrders()

					changesMade = true
					GlobalOrders.SenderId = localID
					fsmChan <- LocalOrders

				} else {
					fmt.Println("None to take: ")
					printOrders()
				}

			default:
				/*if orderIterate(DOWN, N_FLOORS, globalOrdersAvailable){
					changesMade = true
					orderIterate(DOWN, N_FLOORS, scoreAvailableOrder)
					//fmt.Println("Scored")
				} else {
					//fmt.Println("No Orders available")
					//fmt.Println(GlobalOrders.Taken)
				}
				*/

				changesMade = orderIterate(DOWN, N_FLOORS, scoreAvailableOrder)

			}

		case newOrder := <-orderChan:
			switch online {
			case true:
				if newOrder.OrderType != COMMAND {
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
			//changesMade = true

		case newLocalOrders := <-completedOrderChan:
			LocalOrders.Completed = newLocalOrders.Completed
			switch online {
			case true:
				changesMade = orderIterate(COMMAND, N_FLOORS, completeGlobalOrders)
				GlobalOrders.SenderId = localID

			case false:
				changesMade = orderIterate(COMMAND, N_FLOORS, completeLocalOrders)
				fsmChan <- LocalOrders
			}

		case timeOut := <-timeoutChan:
			switch timeOut {
			case "LOCAL_TIMEOUT":
				BackupRestart()
			case "GLOBAL_TIMEOUT":
				changesMade = true
			default:
				fmt.Println("Error timeout?")
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

func addNewLocalOrder(order heis.Order) bool {
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

func addNewGlobalOrder(order heis.Order) bool {
	ordertype := order.OrderType
	floor := order.Floor
	//stamp := time.Now()
	switch ordertype {
	case UP, DOWN:
		if !GlobalOrders.Taken[ordertype][floor] {
			GlobalOrders.Available[ordertype][floor] = true

			for _, ID := range activeElevs {
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

type orderLogic func(heis.ElevButtonType, int) bool

//Iterates over specified intevals with a function,
//returns true if a single iteration resulted in true
func orderIterate(oEnd heis.ElevButtonType, fEnd int, orderFun orderLogic) bool {
	result := false
	for o := UP; o <= oEnd; o++ {
		for f := 0; f < fEnd; f++ {
			if orderFun(o, f) {
				result = true
			}
		}
	}
	return result
}

func globalOrdersAvailable(ordertype heis.ElevButtonType, floor int) bool {
	if GlobalOrders.Available[ordertype][floor] {
		return true
	}
	return false
}

//3 x N_FLOORS
func completeLocalOrders(ordertype heis.ElevButtonType, floor int) bool {
	if LocalOrders.Completed[ordertype][floor] {
		LocalOrders.Pending[ordertype][floor] = false
		LocalOrders.Completed[ordertype][floor] = false
		LocalOrders.Timestamps[ordertype][floor] = time.Time{}
		return true
	}
	return false
}

// 3 x N_FLOORS
func completeGlobalOrders(o heis.ElevButtonType, floor int) bool {
	completed := LocalOrders.Completed

	if completed[o][floor] {
		switch o {
		case UP, DOWN:
			GlobalOrders.Available[o][floor] = false
			GlobalOrders.Taken[o][floor] = false
			GlobalOrders.Timestamps[o][floor] = time.Time{}
			LocalOrders.Pending[o][floor] = false
			LocalOrders.Completed[o][floor] = false
			LocalOrders.Timestamps[o][floor] = time.Time{}

			return true

		case COMMAND:
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
func setLights(ordertype heis.ElevButtonType, floor int) bool {
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
		//heis.ElevSetButtonLamp(ordertype, floor, val)
	default:
		fmt.Println("Invalid ordertype in setLights()")
	}
	heis.ElevSetButtonLamp(ordertype, floor, val)
	return true
}

//2 x N_FLOORS
func scoreAvailableOrder(ordertype heis.ElevButtonType, floor int) bool {
	if scores, ok := GlobalOrders.Scores[localID]; ok {
		if GlobalOrders.Available[ordertype][floor] {
			pFloor, dir := LocalOrders.PrevFloor, LocalOrders.Direction
			floorDiff := (floor - pFloor)
			if floorDiff < 0 {
				floorDiff = -floorDiff
			}

			scores[ordertype][floor] = 200 - floorDiff + floorDiff*int(dir)*10 //COST FUNC
			GlobalOrders.Scores[localID] = scores
			//fmt.Println("Actually scored", GlobalOrders.Scores[localID])

			return true
		} else if GlobalOrders.Taken[ordertype][floor] {
			if temp, ok := GlobalOrders.Scores[localID]; ok {
				temp[ordertype][floor] = 0
				GlobalOrders.Scores[localID] = temp
			}
		}
	}
	//fmt.Println("Nothing to score ", GlobalOrders.Scores)

	return false
}

// 2 x N_FLOORS
func isBestScore(ordertype heis.ElevButtonType, floor int) bool {
	// returns false if it finds a better competitor, else returns true.

	for _, elevID := range activeElevs {
		if value, ok := GlobalOrders.Scores[elevID]; ok {
			if elevID != localID {
				if GlobalOrders.Scores[localID][ordertype][floor] < value[ordertype][floor] {

					return false
				} else if GlobalOrders.Scores[localID][ordertype][floor] == 0 {
					return false
				}
			}
		}
	}
	//fmt.Println("Add score: ", value)

	return true
}

//2 x N_FLOORS
func takeGlobalOrder(ordertype heis.ElevButtonType, floor int) bool {
	changesMade := false
	if isBestScore(ordertype, floor) && GlobalOrders.Available[ordertype][floor] {
		fmt.Println("Taking an order")
		GlobalOrders.Available[ordertype][floor] = false
		GlobalOrders.Taken[ordertype][floor] = true
		GlobalOrders.Timestamps[ordertype][floor] = time.Now()
		LocalOrders.Pending[ordertype][floor] = GlobalOrders.Taken[ordertype][floor]
		LocalOrders.Timestamps[ordertype][floor] = GlobalOrders.Timestamps[ordertype][floor]
		for _, elev := range activeElevs {
			if temp, ok := GlobalOrders.Scores[elev]; ok {
				temp[ordertype][floor] = 0
				GlobalOrders.Scores[elev] = temp
			}
		}
		changesMade = true
	}
	//fmt.Println("Taking an order, taken", GlobalOrders.Taken)

	return changesMade
}

//2 x N_FLOORS
func mergeWithLocOrders(ordertype heis.ElevButtonType, floor int) bool {
	// If an order is listed as taken, but this elev has completed it, the order is removed from globalorders
	if GlobalOrders.Taken[ordertype][floor] && LocalOrders.Completed[ordertype][floor] {
		GlobalOrders.Taken[ordertype][floor] = false
		LocalOrders.Completed[ordertype][floor] = false
		return true
	}
	return false
}

func mergeGlobalOrders(GlobalPacketDEC GlobalOrderStruct) {

	scoreBuffer := GlobalPacketDEC.Scores
	scoreBuffer[localID] = GlobalOrders.Scores[localID]
	/*for _, elevID := range(activeElevs) {
		if value, ok := GlobalPacketENC.Scores[elevID]; ok {
			if elevID == localID{
				scoreBuffer[localID] = GlobalOrders.Scores[localID]
			}
		}
			}
	}*/
	GlobalOrders = GlobalPacketDEC
	GlobalOrders.Scores = scoreBuffer

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

func isLocalTimeout(ordertype heis.ElevButtonType, floor int) bool {
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

func isGlobalTimeout(ordertype heis.ElevButtonType, floor int) bool {
	copyOrder := heis.Order{ordertype, floor}
	if GlobalOrders.Taken[ordertype][floor] {
		if time.Since(GlobalOrders.Timestamps[ordertype][floor]) > time.Second*10 {
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

func printOrders() {

	fmt.Println("\nLocalID: ", localID)
	fmt.Println("activeElevs: ", activeElevs)
	fmt.Println("\n                GlobalOrders Available:")
	for i := 0; i < 2; i++ {
		fmt.Println("                  ", GlobalOrders.Available[i])
	}
	fmt.Println("                 GlobalOrders Taken:")
	for i := 0; i < 2; i++ {
		fmt.Println("                  ", GlobalOrders.Taken[i])
	}
	fmt.Println("\n                Scores:")

	for _, elevID := range activeElevs {
		if value, ok := GlobalOrders.Scores[elevID]; ok {

			fmt.Println("                  ID: 	", elevID, value)

		}
	}

	fmt.Println("\n                LocalOrders Pending:")
	for i := 0; i < 3; i++ {
		fmt.Println("                ", LocalOrders.Pending[i])
	}

}

func timeOut(timeoutChan chan string) {
	for {
		time.Sleep(time.Second)
		if orderIterate(COMMAND, N_FLOORS, isLocalTimeout) {
			//fmt.Println("Timeout local", LocalOrders.Timestamps)
			timeoutChan <- "LOCAL_TIMEOUT"
		} else if orderIterate(DOWN, N_FLOORS, isGlobalTimeout) {
			//fmt.Println("Timeout global", GlobalOrders.Timestamps)
			timeoutChan <- "GLOBAL_TIMEOUT"
		}
	}
}

func BackupRestart() {
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

func loadFromBackup(fsmChan chan fsm.LocalOrderState) {
	backupFilePath := os.Args[1]
	data, _ := ioutil.ReadFile(backupFilePath)
	temp := LocalOrders
	err := json.Unmarshal(data, &temp)
	if err == nil {
		LocalOrders = temp
		changesMade = true
		fsmChan <- LocalOrders
	}
}
