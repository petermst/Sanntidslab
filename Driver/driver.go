package Driver

import (
	. "../Queue"
	"fmt"
	//"os"
	"time"
)

const (
	N_BUTTONS = 3
	N_FLOORS  = 4
)

const (
	MOTOR_DOWN int = -1
	MOTOR_IDLE
	MOTOR_UP
)

const (
	DOOR_CLOSE = 0
	DOOR_OPEN  = 1
)

type ButtonIndicator struct {
	floor  int
	button int
	value  int
}

var lastFloorIndicator int = -1

var lampSetChannelMatrix = [N_FLOORS][N_BUTTONS]int{}

func RunDriver(setButtonIndicator chan ButtonIndicator, setMotorDirection <-chan int, startDoorTimer <-chan bool, eventElevatorStuck chan<- bool, eventAtFloor chan<- int, eventDoorTimeout chan<- bool) {

	checkTicker := time.NewTicker(5 * time.Millisecond).C

	elevatorStuckTimer := time.NewTimer(10 * time.Second)
	elevatorStuckTimer.Stop()

	for {

		select {
		case lamp := <-setButtonIndicator:
			ElevSetButtonLamp(lamp.button, lamp.floor, lamp.value)
			lampSetChannelMatrix[lamp.floor][lamp.button] = lamp.value
		case <-checkTicker:
			checkButtonsPressed(setButtonIndicator)
			checkFloorArrival(eventAtFloor)
		case dir := <-setMotorDirection:
			ElevSetMotorDirection(dir)
			if dir != 0{
				elevatorStuckTimer.Reset(10* time.Second)
			}
		case <-startDoorTimer:
			setDoor(DOOR_OPEN)
			doorTimer := time.NewTimer(time.Second * 3)
			
			var update QueueOperation
			update.operation = false
			update.floor = ElevGetFloorSensorSignal()
			update.elevator = 
			updateQueue <- update

			go func() {
				<-doorTimer.C
				setDoor(DOOR_CLOSE)
				eventDoorTimeout <- true
		case <- elevatorStuckTimer.C:
			eventElevatorStuck <- true
			}()
		}
	}
}

func checkButtonsPressed(calcOptimalElevator  chan<- Order) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if (ElevGetButtonSignal(button, floor) == 1) && (lampSetChannelMatrix[floor][button] == 0) {
				calcOptimalElevator <- Order{floor, button}
			}
		}
	}
}

func checkFloorArrival(eventAtFloor chan<- int) {
	curFloorIndicator := ElevGetFloorSensorSignal()
	if curFloorIndicator != -1 && lastFloorIndicator == -1 {
		ElevSetFloorIndicator(curFloorIndicator)
		//eventAtFloor <- curFloorIndicator
	}
	lastFloorIndicator = curFloorIndicator
}

func setDoor(doorOpen int) {
	ElevSetDoorOpenLamp(doorOpen)
}

func InitializeElevator() int {
	ElevInit()
	ElevSetMotorDirection(-1)
	var initfloor int
	for {
		initfloor = ElevGetFloorSensorSignal()
		if initfloor != -1 {
			ElevSetMotorDirection(0)
			ElevSetFloorIndicator(ElevGetFloorSensorSignal())
			lastFloorIndicator = initfloor
			fmt.Println("\nDriver successfully initialized\n")
			return -1
		}
		time.Sleep(10 * time.Millisecond)
	}
}
