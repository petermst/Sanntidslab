package Driver

import (
	//"fmt"
	//"os"
	"time"
)

type motorDirection int

const (
	N_BUTTONS = 3
	N_FLOORS  = 4
)

const (
	MOTOR_DOWN motorDirection = -1
	MOTOR_IDLE
	MOTOR_UP
)

const (
	DOOR_CLOSE = 0
	DOOR_OPEN  = 1
)

type buttonIndicator struct {
	floor  int
	button int
	value  int
}

var lastFloorIndicator int = -1

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

func RunDriver(setButtonLamp <-chan buttonIndicator, setMotorDirection <-chan int, startDoorTimer <-chan bool) {

	checkTicker := time.NewTicker(5 * time.Millisecond).C

	for {
		select {
		case <-checkTicker:
			checkButtonsPressed()

			//HUSK!!! returnerer ElevGetFloorSensorSignal()
		case lamp := <-setButtonLamp:
			ElevSetButtonLamp(lamp.button, lamp.floor, lamp.value)
		case dir := <-setMotorDirection:
			ElevSetMotorDirection(dir)
		case <-startDoorTimer:
			setDoor(DOOR_OPEN)
		case <-doorTimer:
			setDoor(DOOR_CLOSE)
		}
	}
}

func checkButtonsPressed() {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for button := 0; button < NUM_BUTTONS; button++ {
			if ElevGetButtonSignal(button, floor) && !buttonChannelMatrix[floor][button] {
				// Sette calcOptimalElevator-channel
				buttonChannelMatrix[floor][button] = ElevGetButtonSignal(button, floor)
			}
		}
	}
}

func checkFloorArrival() {
	curFloorIndicator = ElevGetFloorSensorSignal()
	if curFloorIndicator != -1 && lastFloorIndicator == -1 {
		ElevSetFloorIndicator(curFloorIndicator)
		//Sette eventAtFloor-channel
	}
	lastFloorIndicator = curFloorIndicator
}

func setDoor(doorOpen int) {
	ElevSetDoorOpenLamp(doorOpen)
	if doorOpen {
		doorTimer := time.NewTimer(time.Second * 3).C
	}
}
