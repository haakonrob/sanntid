package heisdriver

import (
	_ "errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3
const MOTOR_SPEED = 3000

var lampChannelMatrix = [N_FLOORS][N_BUTTONS]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}

var buttonChannelMatrix = [N_FLOORS][N_BUTTONS]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

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
	TIMER_EVENT
)

type Event struct {
	Type EventType
	Val  int
}

var start_timer bool

func ElevStartTimer() {
	start_timer = true
}

func Poller(orders chan Order, events chan Event) {
	var orderArray [N_BUTTONS][N_FLOORS]bool
	var atFloor bool = true
	var stopped bool
	var obstructed bool
	var timing bool
	var timestamp time.Time

	//time.Sleep(time.Millisecond*100)

	for {
		currFloor := ElevGetFloorSensorSignal()
		stopButton := ElevGetStopSignal()
		obstructionSwitch := ElevGetObstructionSignal()

		if currFloor != -1 && !atFloor {
			atFloor = true
			events <- Event{FLOOR_EVENT, currFloor}
		} else if currFloor == -1 && atFloor {
			atFloor = false
		}

		if stopButton && !stopped {
			stopped = true
			events <- Event{STOP_EVENT, 1}
		} else if !stopButton && stopped {
			stopped = false
			events <- Event{STOP_EVENT, 0}
		}

		if obstructionSwitch && !obstructed {
			obstructed = true
			events <- Event{OBSTRUCTION_EVENT, 1}
		} else if !obstructionSwitch && obstructed {
			obstructed = false
			events <- Event{OBSTRUCTION_EVENT, 0}
		}
		if !timing && start_timer {
			start_timer = false
			timing = true
			timestamp = time.Now()

		} else if timing && (time.Since(timestamp) > time.Second*2) {
			timing = false
			events <- Event{TIMER_EVENT, 1}
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

func ElevSetMotorDirection(dirn ElevMotorDirection) {
	if dirn == DIRN_STOP {
		IoWriteAnalog(MOTOR, 0)
	} else if dirn == DIRN_UP {
		IoClearBit(MOTORDIR)
		IoWriteAnalog(MOTOR, MOTOR_SPEED)
	} else if dirn == DIRN_DOWN {
		IoSetBit(MOTORDIR)
		IoWriteAnalog(MOTOR, MOTOR_SPEED)
	} else {
		fmt.Println("ERROR dir.driver")
	}
}

func ElevSetButtonLamp(button ElevButtonType, floor int, on bool) {
	if button > N_BUTTONS || button < 0 || floor < 0 || floor > N_FLOORS {
		fmt.Println("ERROR set lamp.driver")
	}

	if on {
		IoSetBit(lampChannelMatrix[floor][button])
	} else {
		IoClearBit(lampChannelMatrix[floor][button])
	}
}

func ElevSetFloorIndicator(floor int) {
	// Binary encoding. One light must always be on.
	if floor > N_FLOORS || floor < 0 {
		fmt.Println("ERROR floor ind.driver")
	}

	if floor&0x02 > 0 {
		IoSetBit(LIGHT_FLOOR_IND1)
	} else {
		IoClearBit(LIGHT_FLOOR_IND1)
	}

	if floor&0x01 > 0 {
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

func ElevGetButtonSignal(button ElevButtonType, floor int) bool {

	if button > N_BUTTONS || button < 0 || floor < 0 || floor > N_FLOORS {
		fmt.Println("ERROR get lamp.driver")
	}

	return IoReadBit(buttonChannelMatrix[floor][button])
}

func ElevGetFloorSensorSignal() int {
	if IoReadBit(SENSOR_FLOOR1) {
		return 0
	} else if IoReadBit(SENSOR_FLOOR2) {
		return 1
	} else if IoReadBit(SENSOR_FLOOR3) {
		return 2
	} else if IoReadBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

func ElevGetStopSignal() bool {
	return IoReadBit(STOP)
}

func ElevGetObstructionSignal() bool {
	return IoReadBit(OBSTRUCTION)
}

func ElevInit() {

	InitSuccess := IoInit()
	if !InitSuccess {
		fmt.Println("ERROR init.driver")
	}

	for f := 0; f < N_FLOORS; f++ {
		for b := BUTTON_CALL_UP; b < N_BUTTONS; b++ {
			ElevSetButtonLamp(b, f, false)
		}
	}

	ElevSetStopLamp(false)
	ElevSetDoorOpenLamp(false)
	ElevSetFloorIndicator(0x00)
	ElevSetMotorDirection(DIRN_STOP)

	go func() {
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		fmt.Println("\nTermination signal received. Killing motor.")
		for ElevGetFloorSensorSignal() == -1 {
		}
		ElevSetMotorDirection(DIRN_STOP)
	}()

}
