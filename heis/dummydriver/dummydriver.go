package dummydriver

import (
	_ "errors"
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
)

const N_FLOORS = 4
const N_BUTTONS = 3
const MOTOR_SPEED = 3000

type MotorDirection int
const (
	DIRN_DOWN MotorDirection = -1 << iota
	DIRN_STOP
	DIRN_UP
)

type ButtonType int
const (
	BUTTON_CALL_UP ButtonType = iota
	BUTTON_CALL_DOWN
	BUTTON_COMMAND
)

// Floor 1 is saved as 0, floor 2 is saved as 1, etc
type Order struct {
	Type ButtonType
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

func StartTimer() {
	start_timer = true
}

func Poller(orders chan Order, events chan Event) {
	var orderArray [N_BUTTONS][N_FLOORS]bool
	var atFloor bool = true
	var stopped bool
	var obstructed bool
	var timing bool
	var timestamp time.Time

	for {
		
		currFloor := 1
		stopButton := false
		obstructionSwitch := false
		
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
			events<- Event{TIMER_EVENT,1}
		}
		for i := 0; i < N_FLOORS-1; i++ {
			press := GetButtonSignal(BUTTON_CALL_UP, i)
			if !orderArray[0][i] && press {
				orderArray[0][i] = true
				orders <- Order{BUTTON_CALL_UP, i}
			} else if orderArray[0][i] && !press {
				orderArray[0][i] = false
			}
		}
		for i := 1; i < N_FLOORS; i++ {
			press := GetButtonSignal(BUTTON_CALL_DOWN, i)
			if !orderArray[1][i] && press {
				orderArray[1][i] = true
				orders <- Order{BUTTON_CALL_DOWN, i}
			} else if orderArray[1][i] && !press {
				orderArray[1][i] = false
			}
		}
		for i := 0; i < N_FLOORS; i++ {
			press := GetButtonSignal(BUTTON_COMMAND, i)
			if !orderArray[2][i] && press {
				orderArray[2][i] = true
				orders <- Order{BUTTON_COMMAND, i}
			} else if orderArray[2][i] && !press {
				orderArray[2][i] = false
			}
		}
	}
}

func SetMotorDirection(dirn MotorDirection) {
	fmt.Print(dirn)
}

func SetButtonLamp(button ButtonType, floor int, on bool) {
	fmt.Print(button, floor, on)
}

func SetFloorIndicator(floor int) {
	fmt.Print(floor)
}

func SetDoorOpenLamp(value bool) {
	fmt.Print(value)
}

func SetStopLamp(value bool) {
	fmt.Print(value)
}

func GetButtonSignal(button ButtonType, floor int) bool {
	return false
}

func GetFloorSensorSignal() int {
	return -1
}

func GetStopSignal() bool {
	return false
}

func GetObstructionSignal() bool {
	return false
}

func Init() {
	
	go func (){
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		fmt.Println("\nTermination signal received. Killing motor.")
		for GetFloorSensorSignal() == -1 {}
		SetMotorDirection(DIRN_STOP)
		
		/*backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run coordinator.go ./backupdata")
		backup.Run()*/
		//os.Exit(0)
	}()

}
