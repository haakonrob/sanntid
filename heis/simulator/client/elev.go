package client

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elev.h"
*/
import "C"


const N_FLOORS = 4
const N_BUTTONS = 3
const MOTOR_SPEED = 2800


type ElevMotorDirection int
const (
	DIRN_DOWN ElevMotorDirection = -1 << iota
	DIRN_STOP
	DIRN_UP
)

type ElevButtonType int
const (
	BUTTON_CALL_UP ElevButtonType = iota
	BUTTON_CALL_DOWN
	BUTTON_COMMAND
)

// Floor 1 is saved as 0, floor 2 is saved as 1, etc
type Order struct {
	OrderType ElevButtonType
	Floor     int
}

type EventType int
const (
	NOTHING = iota
	FLOOR_EVENT
	STOP_EVENT
	OBSTRUCTION_EVENT
)

type Event struct {
	Type EventType
	Val  int
}

type ElevType int
const (
	ET_Comedi ElevType = iota
	ET_Simulation
)


func ElevInit(e ElevType){
	C.elev_init(C.elev_type(e))
}

func ElevSetMotorDirection(dirn ElevMotorDirection){
	C.elev_set_motor_direction(C.elev_motor_direction_t(dirn))
}

func ElevSetButtonLamp(button ElevButtonType, floor int, value int){
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}
func ElevSetFloorIndicator(int floor){
	C.elev_set_floor_indicator(C.int(floor))
}
func ElevSetDoorOpenLamp(int value){
	C.elev_set_door_open_lamp(C.int(value))
}
func ElevSetStopLamp(int value){
	C.elev_set_stop_lamp(C.int(value))
}

func ElevGetButtonSignal(button ElevButtonType, floor int)int{
	return int(C.elev_get_button_signal(C.elev_button_type_t(button)), C.int(floor))
}
func ElevGetFloorSensorSignal()int{
	return int(C.elev_get_floor_sensor_signal())
}
func ElevGetStopSignal()int{
	return int(C.elev_get_stop_signal())
}
func ElevGetObstructionSignal()int{
	return int(C.elev_get_obstruction_signal())
}


func Poller(orders chan Order, events chan Event) {
	var orderArray [N_BUTTONS][N_FLOORS]bool
	var atFloor bool = true
	var stopped bool
	var obstructed bool
	
	for {
		currFloor := ElevGetFloorSensorSignal()
		if currFloor != -1 && !atFloor {
			atFloor = true
			events <- Event{FLOOR_EVENT, currFloor}
		} else if currFloor == -1 && atFloor {
			atFloor = false
		}

		stopButton := ElevGetStopSignal()
		if stopButton && !stopped {
			stopped = true
			events <- Event{STOP_EVENT, 1}
		} else if !stopButton && stopped {
			stopped = false
			events <- Event{STOP_EVENT, 0}
		}

		obstructionSwitch := ElevGetObstructionSignal()
		if obstructionSwitch && !obstructed {
			obstructed = true
			events <- Event{OBSTRUCTION_EVENT, 1}
		} else if !obstructionSwitch && obstructed {
			obstructed = false
			events <- Event{OBSTRUCTION_EVENT, 0}
		}

		for i := 0; i < N_FLOORS-1; i++ {
			press := ElevGetButtonSignal(BUTTON_CALL_UP, i)
			if !orderArray[0][i] && press {
				orderArray[0][i] = true
				orders <- Order{BUTTON_CALL_UP, i}
			} else if orderArray[0][i] && !press {
				orderArray[0][i] = false
			}
		}
		for i := 1; i < N_FLOORS; i++ {
			press := ElevGetButtonSignal(BUTTON_CALL_DOWN, i)
			if !orderArray[1][i] && press {
				orderArray[1][i] = true
				orders <- Order{BUTTON_CALL_DOWN, i}
			} else if orderArray[1][i] && !press {
				orderArray[1][i] = false
			}
		}
		for i := 0; i < N_FLOORS; i++ {
			press := ElevGetButtonSignal(BUTTON_COMMAND, i)
			if !orderArray[2][i] && press {
				orderArray[2][i] = true
				orders <- Order{BUTTON_COMMAND, i}
			} else if orderArray[2][i] && !press {
				orderArray[2][i] = false
			}
		}
	}
}