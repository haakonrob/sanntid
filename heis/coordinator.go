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
	LocalOrders  fsm.Orders
	online       bool
	localID      string
	activeElevs  []string
	updateFlag   bool
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
	//GlobalOrders.Scores = make(map[string][2][N_FLOORS]int)

	orderChan := make(chan heis.Order, 5)
	completedOrdersCh := make(chan fsm.Orders)
	eventChan := make(chan heis.Event, 5)
	fsmChan := make(chan fsm.Orders)
	networkCh := make(chan network.Status)

	incomingCh := make(chan []byte)
	outgoingCh := make(chan []byte)
	//incomingCh := make(chan interface{})
	//outgoingCh := make(chan interface{})
	timeoutChan := make(chan string)

	heis.ElevInit()
	go fsm.Fsm(eventChan, fsmChan, completedOrdersCh)
	go network.Monitor(networkCh, loopBack, "sanntidsal", incomingCh, outgoingCh)
	go heis.Poller(orderChan, eventChan)
	//go timeOut(timeoutChan)

	if len(os.Args) > 1 {
		LoadFromBackup(fsmChan)
	}

	tickChan := time.NewTicker(time.Millisecond * 500).C

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

		case <-tickChan:
			printOrders()

			if updateFlag {
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
				updateFlag = false
				resetMask()
			}

		case jmsg := <-incomingCh:
			msg := Message{}
			_ = json.Unmarshal(jmsg, &msg)

			switch msg.ViewCount {
			case 0:
				applyMaskToGlobalOrders(msg.Content)
				msg.Scores[localID] = scoreOrders()

				if msg.SenderID == localID {
					msg.ViewCount++
				}
				jmsg, _ := json.Marshal(msg)
				outgoingCh <- jmsg

			case 1:
				assignOrders(msg.Scores)
				if msg.SenderID != localID {
					jmsg, _ := json.Marshal(msg)
					outgoingCh <- jmsg
				}

			default:
				fmt.Println("Received invalid message")
			}

		case timeOut := <-timeoutChan:
			updateFlag = true
			switch timeOut {
			case "LOCAL_TIMEOUT":
				//BackupRestart()
			case "GLOBAL_TIMEOUT":
				updateFlag = true
				fmt.Println("Global timeout?")
			default:
				fmt.Println("Error timeout?")
			}

			/*
				data, _ := json.Marshal(LocalOrders)
				_ = ioutil.WriteFile("./backupdata", data, 0644)
				if online {
					for i := 0; i < 1; i++ {
						outgoingCh <- EncodeGlobalPacket()
					}
				}
				//			}
			*/
		}
	}
}

/*************************************************************************************************
 Functions for handling and merging LocalOrders and GlobalOrders.
 Most of the complexity is due to the array iteration.
scp ./coordinator student@129.241.187.151:~/coordinator
*************************************************************************************************/

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
			updateFlag = true
		}
	case COMMAND:
		if !LocalOrders.Completed[o][f] {
			LocalOrders.Pending[o][f] = true
		}
	default:
		fmt.Println("Cannot add an invalid order.")
	}
}

func markAsCompleted(completedOrders fsm.Orders) {
	LocalOrders.Pending = completedOrders.Pending
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			LocalOrders.Completed[o][f] = false
			if o != 2 {
				if completedOrders.Completed[o][f] {
					GlobalMask.Taken[o][f] = false
					updateFlag = true
				}
			}
		}
	}
}

func scoreOrders() [2][N_FLOORS]int {
	scores := [2][N_FLOORS]int{}
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			floorDiff := (fsm.GetPrevFloor() - f)
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

func assignOrders(scores map[string][2][N_FLOORS]int) {
	for o := 0; o < 2; o++ {
		for f := 0; f < N_FLOORS; f++ {
			if GlobalOrders.Available[o][f] {
				maxID, max := "", 0
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
				}
			}
		}
	}
}

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

				GlobalOrders.Available[ordertype][floor] = true
				GlobalOrders.Taken[ordertype][floor] = false
				GlobalOrders.Timestamps[ordertype][floor] = time.Now()

				copyOrder.Floor = floor
				copyOrder.OrderType = ordertype

				markAsAvailable(copyOrder)
			case false:
				//dontcare

			}
			return true
		}
	}
	return false
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
}

/*
func timeOut(timeoutChan chan string) {
	for {
		time.Sleep(time.Second)
		if iterateThroughOrders(COMMAND, N_FLOORS, isLocalTimeout) {
			fmt.Println("Timeout local")
			timeoutChan <- "LOCAL_TIMEOUT"
		} else if iterateThroughOrders(DOWN, N_FLOORS, isGlobalTimeout) {
			fmt.Println("Timeout global")
			timeoutChan <- "GLOBAL_TIMEOUT"
		}
	}
}
*/
func BackupRestart() {
	fmt.Println("Local timeout, unknown error.")

	/*
		Avoid timeouting again on restart.
	*/
	for o := UP; o <= COMMAND; o++ {
		for f := 0; f < N_FLOORS; f++ {
			LocalOrders.Timestamps[o][f] = time.Now()
		}
	}

	/*
		Run in a new terminal.
	*/
	backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run coordinator.go ./backupdata")
	backup.Run()

	data, _ := json.Marshal(LocalOrders)
	_ = ioutil.WriteFile("./backupdata", data, 0644)
	os.Exit(0)

}

func LoadFromBackup(fsmChan chan fsm.Orders) {
	backupFilePath := os.Args[1]
	data, _ := ioutil.ReadFile(backupFilePath)
	temp := LocalOrders
	err := json.Unmarshal(data, &temp)
	if err == nil {
		LocalOrders = temp
		updateFlag = true
		fsmChan <- LocalOrders
	}
}
