package fsm

import (
	driver "../heisdriver" //"../simulator/client"
	"fmt"
	"time"
)

/******************************************
TODO
Cleanup, move some funcs to driver
Improve order selection alg

******************************************/
const (
	NUM_FLOORS = driver.N_FLOORS
	UP         = driver.BUTTON_CALL_UP
	DOWN       = driver.BUTTON_CALL_DOWN
	COMMAND    = driver.BUTTON_COMMAND
	NONE       = -1
)

type State int

const (
	IDLE_STATE State = iota
	MOVING_STATE
	COMPLETING_ORDER_STATE
	EM_STOP_STATE
)

/*
This variable is used as means of communication between the coordinator and the state machine.
The coordinator sets the Pending array, and fsm sets the completed, orders.PrevFloor and direction variables.
*/
type LocalOrderState struct {
	Pending    [3][NUM_FLOORS]bool
	Completed  [3][NUM_FLOORS]bool
	Timestamps [3][NUM_FLOORS]time.Time
	PrevFloor  int
	Direction  driver.ElevMotorDirection
}

type stateTransition func()

var stateTable = [4][4]stateTransition{
//  NOTHING 	FLOOR		STOP  		OBSTRUCT 	TIMER
	{nextOrder, null, 		EM_stop, 		null, 	null}, 		/*IDLE_STATE*/
	{null, 		null, 		end_EM_stop, 	null, 	null},   		/*MOVING_STATE*/
	{null, 		null, 		EM_stop, 		null, 	closeDoors},       /*COMPLETING_ORDER_STATE*/
	{null, 		newFloor, 	EM_stop, 		null, 	null}}   		/*EM_STOP_STATE*/

var elevState State
var orders LocalOrderState
var newEvent driver.Event
var updateFlag bool
var destinationOrder driver.Order

func Fsm(eventChan chan driver.Event, coordinatorChan chan LocalOrderState) {

	fsmInit()
	destinationOrder.Floor = NONE
	
	for {
		select {
		case newEvent = <-eventChan:
			stateTable[elevState][newEvent.Type]()
			newEvent = driver.Event{driver.NOTHING, 0}

		case newOrders := <-coordinatorChan:
			orders.Pending = newOrders.Pending
			orders.Completed = newOrders.Completed
			stateTable[elevState][newEvent.Type]()

		// perhaps remove
		default:
			time.Sleep(time.Millisecond * 100)
			stateTable[elevState][newEvent.Type]()

		}
	}
}

func null() {
	return
}

func nextOrder() {
	pending := orders.Pending
	completed := orders.Completed
	foundOrder := false

	var nextOrder driver.Order

	if destinationOrder.Floor == NONE {
	Loop:
		for ordertype := COMMAND; ordertype >= UP; ordertype-- {
			for floor := 0; floor < NUM_FLOORS; floor++ {
				if pending[ordertype][floor] && !completed[ordertype][floor] {
					nextOrder = driver.Order{ordertype, floor}
					foundOrder = true
					destinationOrder = nextOrder
					break Loop
				}
			}
		}
	} else {
		foundOrder = true
		nextOrder = destinationOrder
	}

	if foundOrder {
		if nextOrder.Floor == orders.PrevFloor {
			//fmt.Println("next_order() was already at floor")
			completeOrder(orders.PrevFloor)
		} else if nextOrder.Floor < orders.PrevFloor {
			elevMoveDown()
		} else if nextOrder.Floor > orders.PrevFloor {
			elevMoveUp()
		} else {
			fmt.Println("Failure in nextOrder()")
			return
		}
		elevState = MOVING_STATE
	}

}

func completeOrder(floor int) {
	elevState = COMPLETING_ORDER_STATE

	for ordertype := COMMAND; ordertype >= UP; ordertype-- {
		if orders.Pending[ordertype][orders.PrevFloor] {
			orders.Pending[ordertype][orders.PrevFloor] = false
			orders.Completed[ordertype][orders.PrevFloor] = true
		}
	}
	updateFlag = true
	elevStop()
	driver.ElevSetDoorOpenLamp(true)
	driver.ElevStartTimer()
}

/****Prefereably replace with event******/
/*
func doorTimer(ch ) {
	//fmt.Println("opening doors")
	driver.ElevSetDoorOpenLamp(true)
	time.Sleep(time.Second * 3)
	// Preferably replace with an event
	driver.ElevSetDoorOpenLamp(false)
	elevState = IDLE_STATE

}

/*****************************************/

func newFloor(ch Channels) {
	//fmt.Println("new Floor")
	orders.PrevFloor = newEvent.Val
	floor := orders.PrevFloor
	driver.ElevSetFloorIndicator(floor)

	if destinationOrder.Floor != NONE {
		if ShouldStopOnFloor(floor) && destinationOrder.Floor != floor {
			completeOrder(floor, ch)
		} else if destinationOrder.Floor == floor {
			destinationOrder.Floor = NONE
			completeOrder(floor, ch)
		}
	}
}

func EM_stop() {
	if newEvent.Val > 0 {
		elevStop()
		driver.ElevSetDoorOpenLamp(false)
		elevState = EM_STOP_STATE
		//fmt.Println("Emergency stop")
	}
}

func end_EM_stop() {
	if newEvent.Val == 0 {
		elevState = IDLE_STATE
		//fmt.Println("Now idle")
	}
}

func closeDoors(){
	elevState = IDLE_STATE
	driver.ElevSetDoorOpenLamp(false)
}

/*****Consider moving to driver.go*******/

func elevStop() {
	driver.ElevSetMotorDirection(driver.DIRN_STOP)
	orders.Direction = driver.DIRN_STOP
}

func elevMoveUp() {
	driver.ElevSetMotorDirection(driver.DIRN_UP)
	orders.Direction = driver.DIRN_UP
}

func elevMoveDown() {
	driver.ElevSetMotorDirection(driver.DIRN_DOWN)
	orders.Direction = driver.DIRN_DOWN
}

func ShouldStopOnFloor(floor int) bool {
	pending := orders.Pending
	completed := orders.Completed
	dir := orders.Direction

	switch dir {
	case driver.DIRN_DOWN:
		if pending[driver.BUTTON_CALL_DOWN][floor] && !completed[driver.BUTTON_CALL_DOWN][floor] {
			return true
		}
	case driver.DIRN_UP:
		if pending[driver.BUTTON_CALL_UP][floor] && !completed[driver.BUTTON_CALL_UP][floor] {
			return true
		}
	case driver.DIRN_STOP:
		return true
	default:
		fmt.Println("ERROR in ShouldStopOnFloor()")
	}
	return false
}
/*
func doorTimer(timeout chan<- bool, reset <-chan bool) {
	const doorOpenTime = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()

	for {
		select {
		case <-reset:
			timer.Reset(doorOpenTime)

		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}*/

/****************************************/

func fsmInit() {
	// call getFloorSensor(), if undefined, move to a floor
	elevMoveUp()
	for driver.ElevGetFloorSensorSignal() == -1 {
	}
	elevStop()
	orders.PrevFloor = driver.ElevGetFloorSensorSignal()
	elevState = IDLE_STATE
	updateFlag = false
	newEvent = driver.Event{driver.NOTHING, 0}


}
