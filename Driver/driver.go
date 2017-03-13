package Driver

import (
	. "../Def"
	"fmt"
	"time"
)

var lastFloorIndicator int = -1

var lampSetChannelMatrix [N_FLOORS][N_BUTTONS]int

var doorOpen bool = false

func RunDriver(id string, chQD ChannelsQueueDriver, chFD ChannelsFSMDriver) {

	checkTickerCh := time.NewTicker(5 * time.Millisecond).C

	elevatorStuckTimer := time.NewTimer(10 * time.Second)
	elevatorStuckTimer.Stop()

	for {

		select {
		case lamp := <-chQD.SetButtonIndicatorCh:
			ElevSetButtonLamp(lamp.Button, lamp.Floor, lamp.Value)
			lampSetChannelMatrix[lamp.Floor][lamp.Button] = lamp.Value

		case <-checkTickerCh:
			checkButtonsPressed(chQD.CalcOptimalElevatorCh)
			reached := checkFloorArrival(chFD.EventAtFloorCh)
			if reached {
				elevatorStuckTimer.Reset(10 * time.Second)
			}

		case dir := <-chFD.SetMotorDirectionCh:
			ElevSetMotorDirection(dir)
			if dir != 0 {
				elevatorStuckTimer.Reset(10 * time.Second)
			} else {
				elevatorStuckTimer.Stop()
			}

		case <-chQD.IsDoorOpenCh:
			chQD.IsDoorOpenResponseCh <- doorOpen

		case <-chFD.StartDoorTimerCh:
			doorOpen = true
			setDoor(DOOR_OPEN)
			doorTimer := time.NewTimer(time.Second * 3)
			chQD.UpdateQueueCh <- QueueOperation{false, id, ElevGetFloorSensorSignal(), 0}
			go func() {
				<-doorTimer.C
				setDoor(DOOR_CLOSE)
				doorOpen = false
				chFD.EventDoorTimeoutCh <- true
			}()

		case <-elevatorStuckTimer.C:
			chFD.EventElevatorStuckCh <- true
		}
	}
}

func checkButtonsPressed(calcOptimalElevatorCh chan<- Order) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if (ElevGetButtonSignal(button, floor) == 1) && (lampSetChannelMatrix[floor][button] == 0) {
				calcOptimalElevatorCh <- Order{floor, button}
			}
		}
	}
}

func checkFloorArrival(eventAtFloorCh chan<- int) bool {
	curFloorIndicator := ElevGetFloorSensorSignal()

	if curFloorIndicator != -1 && lastFloorIndicator == -1 {
		ElevSetFloorIndicator(curFloorIndicator)
		eventAtFloorCh <- curFloorIndicator
		lastFloorIndicator = curFloorIndicator
		return true
	}

	lastFloorIndicator = curFloorIndicator
	return false
}

func setDoor(doorOpen int) {
	ElevSetDoorOpenLamp(doorOpen)
}

func InitializeElevator() int {
	ElevInit()
	ElevSetMotorDirection(MOTOR_DOWN)
	var initfloor int

	for {
		initfloor = ElevGetFloorSensorSignal()
		if initfloor != -1 {
			ElevSetMotorDirection(MOTOR_IDLE)
			ElevSetFloorIndicator(ElevGetFloorSensorSignal())
			lastFloorIndicator = initfloor
			fmt.Println("\nDriver successfully initialized\n")
			return initfloor
		}
		time.Sleep(10 * time.Millisecond)
	}
}
