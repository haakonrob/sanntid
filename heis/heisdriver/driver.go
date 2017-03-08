package heisdriver

import (
	"fmt"
	_"errors"
)

const N_FLOORS = 4
const N_BUTTONS = 3
const MOTOR_SPEED = 2800

var lampChannelMatrix = [N_FLOORS][N_BUTTONS] int {
    {LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
    {LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
    {LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
    {LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}


var buttonChannelMatrix = [N_FLOORS][N_BUTTONS] int {
    {BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
    {BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}



type ElevMotorDirection int
const (
	DIRN_DOWN  ElevMotorDirection = -1 << iota 
    DIRN_STOP
    DIRN_UP 
)


type elevButtonType int
const ( 
	BUTTON_CALL_UP elevButtonType = iota
    BUTTON_CALL_DOWN
    BUTTON_COMMAND
)

// Floor 1 is saved as 0, floor 2 is saved as 1, etc
type Order struct {
	OrderType elevButtonType
	Floor int
}

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

func Poller(orders chan Order , events chan Event){
	var orderArray [N_BUTTONS][N_FLOORS] bool
	var atFloor bool
	var stopped bool
	var obstructed bool
	ElevInit()
	for {
		currFloor := ElevGetFloorSensorSignal()
		if currFloor != -1 && !atFloor {
			atFloor = true
			events<- Event{FLOOR_EVENT, currFloor}
		} else if currFloor == -1 && atFloor {
			atFloor = false
		}
		
		stopButton := ElevGetStopSignal()
		if stopButton && !stopped {
			stopped = true
			events<- Event{STOP_EVENT, 1}
		} else if !stopButton && stopped {
			stopped = false
			events<- Event{STOP_EVENT, 0}
		}
		
		obstructionSwitch := ElevGetObstructionSignal()
		if obstructionSwitch && !obstructed {
			obstructed = true
			events<- Event{OBSTRUCTION_EVENT, 1}
		} else if !obstructionSwitch && obstructed {
			obstructed = false
			events<- Event{OBSTRUCTION_EVENT, 1}
		}
		
		for i := 0 ; i < N_FLOORS-1 ; i++ {
			press := ElevGetButtonSignal(BUTTON_CALL_UP, i);
			if !orderArray[0][i] && press {
				orderArray[0][i] = true
				orders<- Order{BUTTON_CALL_UP, i}
			} else if orderArray[0][i] && !press {
				orderArray[0][i] = false
			}
		}
		for i := 1 ; i < N_FLOORS ; i++ {
			press := ElevGetButtonSignal(BUTTON_CALL_DOWN, i);
			if !orderArray[1][i] && press {
				orderArray[1][i] = true
				orders<- Order{ BUTTON_CALL_DOWN, i}
			} else if orderArray[1][i] && !press {
				orderArray[1][i] = false
			}
		}
		for i := 0 ; i < N_FLOORS ; i++ {
			press := ElevGetButtonSignal(BUTTON_COMMAND, i);
			if !orderArray[2][i] && press {
				orderArray[2][i] = true
				orders<- Order{ BUTTON_COMMAND, i}
			} else if orderArray[2][i] && !press {
				orderArray[2][i] = false
			}
		}
	}
}


func ElevSetMotorDirection(dirn ElevMotorDirection ) {
    if dirn == DIRN_STOP{
		fmt.Println("STOP")
        IoWriteAnalog(MOTOR, 0)
    } else if dirn == DIRN_UP {
		fmt.Println("UP")
        IoClearBit(MOTORDIR)
        IoWriteAnalog(MOTOR, MOTOR_SPEED)
    } else if dirn == DIRN_DOWN {
		fmt.Println("DOWN")
        IoSetBit(MOTORDIR)
        IoWriteAnalog(MOTOR, MOTOR_SPEED)
    }else {
		fmt.Println("ERROR dir.driver")
	}
}

func ElevSetButtonLamp(button int ,  floor int,  value int) {
    
	if button > N_BUTTONS || button < 0 || floor < 0 || floor > N_FLOORS {
		fmt.Println("ERROR set lamp.driver")
	}
	
	//fmt.Println("Button: ", button, "Floor: ", floor, "Value: ", value)

    if value != 0 {	
        IoSetBit(lampChannelMatrix[floor][button])
    } else {
		//fmt.Println("turnoff")
        IoClearBit(lampChannelMatrix[floor][button])
    }
}

func ElevSetFloorIndicator(floor int) {
    // Binary encoding. One light must always be on.
	if floor > N_FLOORS || floor < 0{
		fmt.Println("ERROR floor ind.driver")	
	}

    if floor&0x02 > 0{
        IoSetBit(LIGHT_FLOOR_IND1)
    } else {
        IoClearBit(LIGHT_FLOOR_IND1)
    }    

    if floor&0x01 > 0{
        IoSetBit(LIGHT_FLOOR_IND2)
    } else {
        IoClearBit(LIGHT_FLOOR_IND2)
    }   
}

func ElevSetDoorOpenLamp(value bool) {
    if value {
        IoSetBit(LIGHT_DOOR_OPEN)
    } else {
        IoClearBit(LIGHT_DOOR_OPEN)
    }
}

func ElevSetStopLamp(value bool) {
    if value {
        IoSetBit(LIGHT_STOP)
    } else {
        IoClearBit(LIGHT_STOP)
    }
}

func ElevGetButtonSignal(button elevButtonType, floor int) bool{

	if button > N_BUTTONS || button < 0 || floor < 0 || floor > N_FLOORS {
		fmt.Println("ERROR get lamp.driver")
	}

    return IoReadBit(buttonChannelMatrix[floor][button])
}

func ElevGetFloorSensorSignal()int {
    if IoReadBit(SENSOR_FLOOR1){
        return 0
    } else if IoReadBit(SENSOR_FLOOR2){
        return 1
    } else if IoReadBit(SENSOR_FLOOR3){
        return 2
    } else if IoReadBit(SENSOR_FLOOR4) {
        return 3
    } else {
        return -1
    }
}

func ElevGetStopSignal()bool {
    return IoReadBit(STOP)
}


func ElevGetObstructionSignal() bool{
    return IoReadBit(OBSTRUCTION)
}

func ElevInit(){

	InitSuccess := IoInit()
	if !InitSuccess{
		fmt.Println("ERROR init.driver")
	}

    for f := 0; f < N_FLOORS; f++ {
        for b := 0; b < N_BUTTONS; b++{            
			ElevSetButtonLamp(b, f, 0)
        }
    }


    ElevSetStopLamp(false)
    ElevSetDoorOpenLamp(false)
    ElevSetFloorIndicator(0x00)
	ElevSetMotorDirection(DIRN_STOP)

}
