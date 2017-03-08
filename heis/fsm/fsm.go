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