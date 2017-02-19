package main

import (
	"fmt"
	def "./heisdriver"
	_"time"
)

func main() {
    def.ElevInit()
  
	


	
    for {
        // Change direction when we reach top/bottom floor

		fmt.Println("Floor: ", def.ElevGetFloorSensorSignal())	
        if (def.ElevGetFloorSensorSignal() == def.N_FLOORS - 1) {
            def.ElevSetMotorDirection(def.DIRN_DOWN)
        } else if (def.ElevGetFloorSensorSignal() == -1) {
            def.ElevSetMotorDirection(def.DIRN_DOWN)
        }

		if(def.ElevGetButtonSignal(1,1)){
			def.ElevSetStopLamp(true)
		}
		
        // Stop elevator and exit program if the stop button is pressed
        /*if (def.ElevGetStopSignal()) {
            def.ElevSetMotorDirection(def.DIRN_STOP)
			fmt.Println("End", def.ElevGetStopSignal())
            //return 0
        }*/
    }

	
}
