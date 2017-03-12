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
)

type State int

const (
	IDLE_STATE State = iota
	STOPPED_CLOSED_STATE
	STOPPED_OPEN_STATE
	MOVING_STATE
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
	//	NOTHING 		FLOOR_EVENT 	STOP_EVENT  	OBSTRUCT_EVENT
	{next_order, null, EM_stop, null}, /*IDLE_STATE*/
	{null, null, end_EM_stop, null},   /*STOPPED_CLOSED_STATE*/
	{null, null, EM_stop, null},       /*STOPPED_OPEN_STATE*/
	{null, newFloor, EM_stop, null}}   /*MOVING_STATE*/

var elevState State
var orders LocalOrderState
var newEvent driver.Event
var updateFlag bool

func Fsm(eventChan chan driver.Event, coordinatorChan chan LocalOrderState) {

	fsmInit()

	for {
		select {
		case newEvent = <-eventChan:
			stateTable[elevState][newEvent.Type]()
			newEvent = driver.Event{driver.NOTHING, 0}

		case newOrders := <-coordinatorChan:

			orders.Pending = newOrders.Pending
			orders.Completed = newOrders.Completed
			stateTable[elevState][newEvent.Type]()

			//fmt.Println(" FSM C pend: \n", orders.Pending)
			//fmt.Println(" \n FSM C comp: \n", orders.Completed)

		default:
			time.Sleep(time.Millisecond * 200)

			//fmt.Println("FSM State: ", elevState)
			//fmt.Println("FSM newEvent: ", newEvent)

			if updateFlag {
				fmt.Println("trying to update coordinator")
				coordinatorChan <- orders
				fmt.Println("Update succesful")
				updateFlag = false
				elevState = IDLE_STATE

				//wait for door to close...
				time.Sleep(time.Second * 4)
				stateTable[elevState][newEvent.Type]()
			}

		}

	}

}

func null() {
	//fmt.Println("null")
	return
}

func next_order() {

	pending := orders.Pending
	completed := orders.Completed
	foundOrder := false
	var nextOrder driver.Order
Loop:
	for ordertype := COMMAND; ordertype >= UP; ordertype-- {
		for floor := 0; floor < NUM_FLOORS; floor++ {
			if pending[ordertype][floor] && !completed[ordertype][floor] {
				nextOrder = driver.Order{ordertype, floor}
				foundOrder = true
				break Loop
			}
		}
	}
	if foundOrder {
		if nextOrder.Floor == orders.PrevFloor {
			//fmt.Println("next_order() was already at floor")
			complete_order(orders.PrevFloor)
		} else if nextOrder.Floor < orders.PrevFloor {
			elev_move_down()
		} else if nextOrder.Floor > orders.PrevFloor {
			elev_move_up()
		} else {
			fmt.Println("Failure in next_order()")
			return
		}
		elevState = MOVING_STATE
	}

}

func complete_order(floor int) {
	elevState = STOPPED_OPEN_STATE
	for ordertype := COMMAND; ordertype >= UP; ordertype-- {
		if orders.Pending[ordertype][orders.PrevFloor] {
			orders.Pending[ordertype][orders.PrevFloor] = false
			orders.Completed[ordertype][orders.PrevFloor] = true
		}
	}

	elevState = STOPPED_OPEN_STATE
	updateFlag = true
	elev_stop()
	//fmt.Println("complete_order:", orders.Completed)

	go doorTimer()
}

/****Prefereably replace with event******/
func doorTimer() {
	//fmt.Println("opening doors")
	driver.ElevSetDoorOpenLamp(true)
	time.Sleep(time.Second * 3)
	// Preferably replace with an event
	driver.ElevSetDoorOpenLamp(false)
	elevState = IDLE_STATE
}

/*****************************************/

func newFloor() {
	//fmt.Println("new Floor")
	orders.PrevFloor = newEvent.Val
	floor := orders.PrevFloor
	driver.ElevSetFloorIndicator(floor)
	pending := orders.Pending
	completed := orders.Completed
	for ordertype := COMMAND; ordertype >= UP; ordertype-- {
		if pending[ordertype][floor] && !completed[ordertype][floor] {
			complete_order(floor)
		}
	}

}

func EM_stop() {
	if newEvent.Val > 0 {
		elev_stop()
		driver.ElevSetDoorOpenLamp(false)
		elevState = STOPPED_CLOSED_STATE
		//fmt.Println("Emergency stop")
	}
}

func end_EM_stop() {
	if newEvent.Val == 0 {
		elevState = IDLE_STATE
		//fmt.Println("Now idle")
	}
}

/*****Consider moving to driver.go*******/

func elev_stop() {
	driver.ElevSetMotorDirection(driver.DIRN_STOP)
	orders.Direction = driver.DIRN_STOP
}

func elev_move_up() {
	driver.ElevSetMotorDirection(driver.DIRN_UP)
	orders.Direction = driver.DIRN_UP

}

func elev_move_down() {
	driver.ElevSetMotorDirection(driver.DIRN_DOWN)
	orders.Direction = driver.DIRN_DOWN

}

/****************************************/

func fsmInit() {
	// call getFloorSensor(), if undefined, move to a floor
	elev_move_up()
	for driver.ElevGetFloorSensorSignal() == -1 {
	}
	elev_stop()
	orders.PrevFloor = driver.ElevGetFloorSensorSignal()
	elevState = IDLE_STATE
	updateFlag = false
	newEvent = driver.Event{driver.NOTHING, 0}
}