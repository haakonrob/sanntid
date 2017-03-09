<<<<<<< HEAD
package main

import (
	"fmt"
	elev "../heisdriver"
)

//DRIVER
//ElevSetMotorDirection(dirn elevMotorDirection )
//ElevSetButtonLamp(button int ,  floor int,  value int)
//ElevSetFloorIndicator(floor int) 
//ElevSetDoorOpenLamp(value bool)
//func ElevSetStopLamp(value bool) {
//func ElevGetButtonSignal(button int, floor int) bool{
//func ElevGetFloorSensorSignal()int {
//func ElevGetStopSignal()bool {
//func ElevGetObstructionSignal() bool{
//finish(timer)

type EventType int
const (
	FLOOR_EVENT = iota
	STOP_EVENT
	OBSTRUCTION_EVENT
)
type Event struct {
	Type EventType
	val int
}


type FsmPackageStruct struct {
	Pending[3][elev.NUM_FLOORS] bool
	Completed[3][elev.NUM_FLOORS] bool
	prevFloor int
    dir elev.elevMotorDirection
}

var coordPack FsmPackageStruct
var evPack Event

func FsmInit()bool{
	//if crash not happened
	if(1){ 

		for f := 0; f < N_FLOORS; f++ {
	        for b := 0; b < 3; b++{            
				coordPack.Pending[b][f] = 0
				coordPack.Completed[b][f] = 0
	        }
	    }
	}else{
		fmt.Println("LOAD FROM FILE")
		//else load from file
	}

	//If lift not at floor
	
	elev.ElevSl();
	coordPack.dir = DIRN_UP;

	 

	//everything to zero:

	//find floors
	elev.ElevSetMotorDirection(elev.DIRN_UP);

	//set state idle
}


func fsm(eventChan chan Event, coordChan chan Struct){
	var sucsess = FsmInit()

	select{
		case orderPack := <-coordChan:
			coordPack.Pending = coordPack.Pending
			coordPack.Completed = coordPack.Completed
					
		case evPack := <- eventChan:
			updatePackage(evPack);

		/*case idleOrder := <-eventChan: 
			//set motor dir
			//turn off lights
			//
		case executeOrder
			//set motor dir after floor
			//set lights
			//
			//
			//timestamp
			//
		case completeOrder
			//turn of light
			//notify coordinator
			//delete order
			//timer wait
		*/
	}

}


func updatePackage(order Order struct){
	liftAss.Pending = order.
	Completed[3][elev.NUM_FLOORS] bool
	prevFloor int
    dir elev.elevMotorDirection

}

func GoToFloor(floor int) bool{

}


	Pending[3][elev.NUM_FLOORS] bool
	Completed[3][elev.NUM_FLOORS] bool
	prevFloor int
    dir elev.elevMotorDirection
=======
package fsm

import (
	"fmt"
	"time"
	driver "../heisdriver"
)

const (
	NUM_FLOORS = driver.N_FLOORS
	UP = driver.BUTTON_CALL_UP
	DOWN = driver.BUTTON_CALL_DOWN
	COMMAND = driver.BUTTON_COMMAND
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
	Pending [3][NUM_FLOORS] bool
	Completed [3][NUM_FLOORS] bool
	PrevFloor int
	Direction driver.ElevMotorDirection
}

type stateTransition func()
var stateTable = [4][4]stateTransition{
//	NOTHING 		FLOOR_EVENT 	STOP_EVENT  	OBSTRUCT_EVENT
	{next_order, 	null, 			EM_stop, 		null}, 		/*IDLE_STATE*/
	{null, 			null, 			end_EM_stop, 	null},   	/*STOPPED_CLOSED_STATE*/
	{null, 			null, 			EM_stop, 		null},      /*STOPPED_OPEN_STATE*/
	{null, 			newFloor,		EM_stop, 		null}}      /*MOVING_STATE*/

var elevState State
var orders LocalOrderState
var newEvent driver.Event
var updateFlag bool

func Fsm(eventChan chan driver.Event, coordinatorChan chan LocalOrderState) {
	fsmInit()
	for {
		select {
			case newEvent = <-eventChan:
				fmt.Println("Event:", newEvent)
				stateTable[elevState][newEvent.Type]()
				newEvent = driver.Event{driver.NOTHING, 0}
			case newOrders := <-coordinatorChan:
				orders.Pending = newOrders.Pending
				orders.Completed = newOrders.Completed
				fmt.Println("C pend:", orders.Pending)
				fmt.Println("C comp:", orders.Completed)
			default:
				if updateFlag {
					fmt.Println("trying to update coordinator")
					coordinatorChan<-orders
					fmt.Println("Update succesful")
					updateFlag = false
				}
				if elevState == IDLE_STATE {
					stateTable[elevState][newEvent.Type]()
					newEvent = driver.Event{driver.NOTHING, 0}
				}
		}
		
	}

}

func fsmInit() {
	// call getFloorSensor(), if undefined, move to a floor
	elev_move_up()
	for driver.ElevGetFloorSensorSignal() == -1 {}
	elev_stop()
	orders.PrevFloor = driver.ElevGetFloorSensorSignal()
	elevState = IDLE_STATE
	updateFlag = false
	newEvent = driver.Event{driver.NOTHING, 0}
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
		for ordertype:= COMMAND;ordertype>=UP;ordertype--{
			for floor:=0;floor<NUM_FLOORS;floor++{
				if pending[ordertype][floor] && !completed[ordertype][floor] {
					nextOrder = driver.Order{ordertype,floor}
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
	for ordertype := COMMAND;ordertype>=UP;ordertype--{
		if orders.Pending[ordertype][orders.PrevFloor]{
			orders.Pending[ordertype][orders.PrevFloor] = false
			orders.Completed[ordertype][orders.PrevFloor] = true
		}
	}
	elevState = STOPPED_OPEN_STATE
	updateFlag = true
	elev_stop()
	fmt.Println("complete_order:", orders.Completed)
	go doorTimer()

	
}

func doorTimer(){
		driver.ElevSetDoorOpenLamp(true)
		time.Sleep(time.Second*3)
		// Preferably replace with an event
		driver.ElevSetDoorOpenLamp(false)
		elevState = IDLE_STATE
	}

func newFloor() {
	fmt.Println("new Floor")
	orders.PrevFloor = newEvent.Val
	floor := orders.PrevFloor
	driver.ElevSetFloorIndicator(floor)
	pending := orders.Pending
	completed := orders.Completed
	for ordertype := COMMAND;ordertype>=UP;ordertype--{
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
		fmt.Println("Emergency stop")
	}
}

func end_EM_stop(){
	if newEvent.Val == 0 {
		elevState = IDLE_STATE
		fmt.Println("Now idle")
	}
}

func elev_stop(){
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
>>>>>>> 78f8e02a8fbc3791b0e64b10b75e0a0c20bff8e0
