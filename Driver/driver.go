package Driver

import (
	. "../Def"
	"fmt"
	//"os"
	"time"
)

var lastFloorIndicator int = -1

var lampSetChannelMatrix [N_FLOORS][N_BUTTONS]int

var doorOpen bool = false

func RunDriver(id string, setButtonIndicatorCh chan ButtonIndicator, setMotorDirectionCh <-chan int, startDoorTimerCh <-chan bool, eventElevatorStuckCh chan<- bool, eventAtFloorCh chan<- int, eventDoorTimeoutCh chan<- bool, calcOptimalElevatorCh chan<- Order, updateQueueCh chan<- QueueOperation, isDoorOpenCh <-chan bool, isDoorOpenResponseCh chan<- bool) {

	checkTickerCh := time.NewTicker(5 * time.Millisecond).C

	elevatorStuckTimer := time.NewTimer(10 * time.Second)
	elevatorStuckTimer.Stop()

	for {
		//fmt.Println(lampSetChannelMatrix)
		select {
		case lamp := <-setButtonIndicatorCh:
			ElevSetButtonLamp(lamp.Button, lamp.Floor, lamp.Value)
			lampSetChannelMatrix[lamp.Floor][lamp.Button] = lamp.Value
		case <-checkTickerCh:
			checkButtonsPressed(calcOptimalElevatorCh)
			reached := checkFloorArrival(eventAtFloorCh)
			if reached {
				elevatorStuckTimer.Reset(10 * time.Second)
			}
		case dir := <-setMotorDirectionCh:
			ElevSetMotorDirection(dir)
			if dir != 0 {
				elevatorStuckTimer.Reset(10 * time.Second)
			} else {
				elevatorStuckTimer.Stop()
			}
		case <-isDoorOpenCh:
			isDoorOpenResponseCh <- doorOpen
		case <-startDoorTimerCh:
			doorOpen = true

			setDoor(DOOR_OPEN)
			doorTimer := time.NewTimer(time.Second * 3)

			updateQueueCh <- QueueOperation{false, id, ElevGetFloorSensorSignal(), 0}

			go func() {
				<-doorTimer.C
				setDoor(DOOR_CLOSE)
				doorOpen = false
				eventDoorTimeoutCh <- true
			}()

		case <-elevatorStuckTimer.C:
			eventElevatorStuckCh <- true
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
