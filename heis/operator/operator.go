package elevoperator

import (
	
	"fmt"
	elev "../dummydriver" //"../simulator/client"
)

/******************************************
TODO
Cleanup, move some funcs to elev
Improve order selection alg

******************************************/
const NONE = -1

type state int
const (
	IDLE_STATE state = iota
	MOVING_STATE
	COMPLETING_ORDER_STATE
	EM_STOP_STATE
)

type elevState  struct {
	State state
	PrevFloor int
	Direction elev.MotorDirection
}

type stateTransition func()
var stateTable = [4][5]stateTransition{
//  NOTHING 	FLOOR		STOP  		OBSTRUCT 	TIMER
	{getNextOrder,	null, 	EM_stop, 	null, 	null}, 		/*IDLE_STATE*/
	{null, 	  newFloor, 	EM_stop, 	null, 	null},   	/*MOVING_STATE*/
	{null, 		null, 		EM_stop, 	null, 	closeDoors},       /*COMPLETING_ORDER_STATE*/
	{null, 		null, 	  end_EM_stop,	null, 	null}}   		/*EM_STOP_STATE*/



var Elev elevState  
var newEvent elev.Event
var nextOrder elev.Order
var pendingOrders []elev.Order
var updateFlag bool
var completedOrders chan elev.Order



func Init() {
	// call getFloorSensor(), if undefined, move to a floor
	elev.SetMotorDirection(elev.DIRN_UP)
	for elev.GetFloorSensorSignal() == -1 {}
	elev.SetMotorDirection(elev.DIRN_STOP)
	Elev.PrevFloor = elev.GetFloorSensorSignal()
	elev.SetFloorIndicator(Elev.PrevFloor)
	Elev.State = IDLE_STATE
	newEvent = elev.Event{elev.NOTHING, 0}	
}

func Start(eventChan chan elev.Event, coordinatorChan <-chan elev.Order, completedOrderChan chan elev.Order) {
	Init()
	completedOrders = completedOrderChan
	nextOrder.Floor = NONE
	for {
		select {
		case newEvent = <-eventChan:
			stateTable[Elev.State][newEvent.Type]()
			newEvent = elev.Event{elev.NOTHING, 0}

		case newOrder := <-coordinatorChan:
			pendingOrders = append(pendingOrders, newOrder)
		}

		//nextOrder() is triggered by a nothing event
		stateTable[Elev.State][newEvent.Type]()
	}
}

/*
	State transition functions
*/

func null() {
	return
}


func EM_stop() {
	if newEvent.Val > 0 {
		ElevStop()
		elev.SetDoorOpenLamp(false)
		Elev.State = EM_STOP_STATE
	}
}

func end_EM_stop() {
	if newEvent.Val == 0 {
		Elev.State  = IDLE_STATE
	}
}

func closeDoors(){
	Elev.State  = IDLE_STATE
	elev.SetDoorOpenLamp(false)
}


func getNextOrder() {
	if len(pendingOrders) == 0 || nextOrder.Floor != NONE {
		return
	}
	priority, max_i := 0, 0
	if nextOrder.Floor == NONE  {
		for i, order := range(pendingOrders){
			if val := (50+ int(order.Type)*25 - 10*i)  ; val > priority {
				priority = val
				max_i = i
			}
		}
	}
	nextOrder = pendingOrders[max_i]
	if nextOrder.Floor == elev.GetFloorSensorSignal() {
		Complete(nextOrder)
	} else if nextOrder.Floor < Elev.PrevFloor {
		ElevMoveDown()
		Elev.State  = MOVING_STATE
	} else if nextOrder.Floor > Elev.PrevFloor {
		ElevMoveUp()
		Elev.State  = MOVING_STATE
	} else {
		fmt.Println("Failure in nextOrder()")
		return
	}

}

func newFloor() {
	Elev.PrevFloor = newEvent.Val
	floor := newEvent.Val
	elev.SetFloorIndicator(floor)

	fakeOrderUp := elev.Order{elev.BUTTON_CALL_UP, floor}
	fakeOrderDown := elev.Order{elev.BUTTON_CALL_DOWN, floor}
	if nextOrder.Floor != NONE {
		if CanCompleteOrders(floor) && nextOrder.Floor != floor {
			Complete(fakeOrderUp)
			Complete(fakeOrderDown)
		} else if nextOrder.Floor == floor {
			nextOrder.Floor = NONE
			Complete(fakeOrderUp)
			Complete(fakeOrderDown)
		}
	}
}


/*
	Order handling functions
*/

func findOrder(order elev.Order, list []elev.Order)(int, bool){
	i, found := 0, false
	for j, pendingOrder := range(list){
		if found = (order == pendingOrder) ; found {
			i = j
		}
	}
	return i, found
}

func removeOrder(order elev.Order, list []elev.Order)([]elev.Order, bool){
	oldList := list
	if j, ok := findOrder(order, list) ; ok {
		copy(list[j:], list[j+1:])
		list[len(list)-1] = elev.Order{}
		return list[:len(list)-1], true
	}
	return oldList, false
}


func Complete(order elev.Order) {
	temp, ok := removeOrder(order, pendingOrders)
	if ok {
		pendingOrders = temp
		Elev.State = COMPLETING_ORDER_STATE
		ElevStop()
		elev.SetDoorOpenLamp(true)
		elev.StartTimer()
		if nextOrder.Floor == order.Floor {
			nextOrder.Floor = NONE
		}
		completedOrders <- order
	}
}

func CanCompleteOrders(floor int) bool {
	if _, ok := findOrder(elev.Order{elev.BUTTON_CALL_DOWN, floor}, pendingOrders) ; ok {
		return true
	}
	switch Elev.Direction {
	case elev.DIRN_DOWN:
		_, ok := findOrder(elev.Order{elev.BUTTON_CALL_DOWN, floor}, pendingOrders) 
		return ok
	case elev.DIRN_UP:
		_, ok := findOrder(elev.Order{elev.BUTTON_CALL_UP, floor}, pendingOrders) 
		return ok
	case elev.DIRN_STOP:
		return true
	default:
		fmt.Println("ERROR in ShouldStopOnFloor()")
	}
	return false
}

func ScoreOrder(order elev.Order)(int){
	prevFloor, dir := Elev.PrevFloor, Elev.Direction
	floorDiff := (order.Floor - prevFloor)	
	if floorDiff < 0 { floorDiff = -floorDiff }
	//randNum := int(byte(localID[len(localID)-1]))
	return 100-floorDiff*20 + (order.Floor - prevFloor)*int(dir)*10 
}


/*****Consider moving to elev.go*******/

func ElevStop() {
	elev.SetMotorDirection(elev.DIRN_STOP)
	Elev.Direction = elev.DIRN_STOP
}

func ElevMoveUp() {
	elev.SetMotorDirection(elev.DIRN_UP)
	Elev.Direction = elev.DIRN_UP
}

func ElevMoveDown() {
	elev.SetMotorDirection(elev.DIRN_DOWN)
	Elev.Direction = elev.DIRN_DOWN
}
/****************************************/

