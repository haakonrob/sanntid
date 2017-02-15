package heisdriver

import (
	"fmt"
	"./channels"
	"./io"
)

func main() {
    elevInit();
    elev_set_motor_direction(DIRN_UP);


    for {
        // Change direction when we reach top/bottom floor
        if (elevGetFloorSensorSignal() == N_FLOORS - 1) {
            elevSetMotorDirection(DIRN_DOWN);
        } else if (elevGetFloorSensorSignal() == 0) {
            elevSetMotorDirection(DIRN_UP);
        }

        // Stop elevator and exit program if the stop button is pressed
        if (elevGetStopSignal()) {
            elevSetMotorDirection(DIRN_STOP);
            return 0;
        }
    }
}