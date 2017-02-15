package driver

import {
	"fmt"
	_"errors"
}

const lampChannelMatrix = [N_FLOORS][N_BUTTONS]int(
    {LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
    {LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
    {LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
    {LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
)


const buttonChannelMatrix = [N_FLOORS][N_BUTTONS]int(
    {BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
    {BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
)

const N_FLOORS 4
const N_BUTTONS 3

type tagElevMotorDirection int
type tagElevLampType int


const(
	DIRN_DOWN  = -1,
    DIRN_STOP  = 0,
    DIRN_UP  = 1
)tagElevMotorDirection

const( 
	BUTTON_CALL_UP = 0,
    BUTTON_CALL_DOWN = 1,
    BUTTON_COMMAND = 2
)tagElevLampType



func elevInit(){

	bool initSuccess = ioInit();

    //assert(init_success && "Unable to initialize elevator hardware!");

    for f := 0; f < N_FLOORS; f++ {
        for b elevButtonType = 0; b < N_BUTTONS; b++{
            elevSetButtonLamp(b, f, 0)
        }
    }

    elevSetStopLamp(0);
    elevSetDoorOpenLamp(0);
    elevSetFloorIndicator(0);

}

func elevSetMotorDirection(dirn elevMotorDirection ) {
    if dirn == 0{
        ioWriteAnalog(MOTOR, 0);
    } else if dirn > 0 {
        ioClearBit(MOTORDIR);
        ioWriteAnalog(MOTOR, MOTOR_SPEED);
    } else if dirn < 0 {
        ioSetBit(MOTORDIR);
        ioWriteAnalog(MOTOR, MOTOR_SPEED);
    }
}

func elevSetButtonLamp(button elevButtonType ,  floor int,  value int) {
    //assert(floor >= 0);
    //assert(floor < N_FLOORS);
    //assert(button >= 0);
    //assert(button < N_BUTTONS);

    if value {
        ioSetBit(lampChannelMatrix[floor][button]);
    } else {
        ioClearBit(lampChannelMatrix[floor][button]);
    }
}

func elevSetFloorIndicator(floor int) {
    //assert(floor >= 0);
    //assert(floor < N_FLOORS);

    // Binary encoding. One light must always be on.
    if floor & 0x02 {
        ioSetBit(LIGHT_FLOOR_IND1);
    } else {
        ioClearBit(LIGHT_FLOOR_IND1);
    }    

    if floor & 0x01 {
        ioSetBit(LIGHT_FLOOR_IND2);
    } else {
        ioClearBit(LIGHT_FLOOR_IND2);
    }    
}

func elevSetDoorOpenLamp(value int) {
    if value {
        ioSetBit(LIGHT_DOOR_OPEN);
    } else {
        ioClearBit(LIGHT_DOOR_OPEN);
    }
}

func elevSetStopLamp(value int) {
    if value {
        ioSetBit(LIGHT_STOP);
    } else {
        ioClearBit(LIGHT_STOP);
    }
}

func elev_get_button_signal(button elevButtonType, floor int) int{
    //assert(floor >= 0);
    //assert(floor < N_FLOORS);
    //assert(button >= 0);
    //assert(button < N_BUTTONS);


    return ioReadBit(buttonChannelMatrix[floor][button]);
}

func elevGetFloorSensorSignal(void)int {
    if ioReadBit(SENSOR_FLOOR1) {
        return 0;
    } else if ioReadBit(SENSOR_FLOOR2) {
        return 1;
    } else if ioReadBit(SENSOR_FLOOR3) {
        return 2;
    } else if ioReadBit(SENSOR_FLOOR4) {
        return 3;
    } else {
        return -1;
    }
}

func elevGetStopSignal(void)int {
    return ioReadBit(STOP);
}


func elev_get_obstruction_signal(void) int{
    return ioReadBit(OBSTRUCTION);
}