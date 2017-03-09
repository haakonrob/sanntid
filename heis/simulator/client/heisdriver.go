package heisdriver

import (
	"fmt"
	_"errors"
)

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


func ElevSetMotorDirection(dirn elevMotorDirection ) {
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

func ElevGetButtonSignal(button int, floor int) bool{

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
