package Driver

import (
	"fmt"
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

type ButtonIndicator struct {
	floor  int
	button int
	value  int
}

var lastFloorIndicator int = -1

var lampSetChannelMatrix = [N_FLOORS][N_BUTTONS]int{}

func RunDriver(setButtonIndicator chan ButtonIndicator, setMotorDirection <-chan int, startDoorTimer <-chan bool) {

	checkTicker := time.NewTicker(5 * time.Millisecond).C

	for {
		select {
		case <-checkTicker:
			checkButtonsPressed(setButtonIndicator)

			//HUSK!!! returnerer ElevGetFloorSensorSignal()
		case lamp := <-setButtonIndicator:
			fmt.Println("et eller anna")
			ElevSetButtonLamp(lamp.button, lamp.floor, lamp.value)
		case dir := <-setMotorDirection:
			ElevSetMotorDirection(dir)
		case <-startDoorTimer:
			setDoor(DOOR_OPEN)
			doorTimer := time.NewTimer(time.Second * 3)
			go func() {
				<-doorTimer.C
				setDoor(DOOR_CLOSE)
			}()
		}
	}
}

func checkButtonsPressed(setButtonIndicator chan<- ButtonIndicator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if (ElevGetButtonSignal(button, floor) == 1) && (lampSetChannelMatrix[floor][button] == 0) {
				// Sette calcOptimalElevator-channel
				lampSetChannelMatrix[floor][button] = ElevGetButtonSignal(button, floor)

				var but ButtonIndicator
				but.floor = floor
				but.button = button
				but.value = 1
				fmt.Printf("%d\n", but.button)
				setButtonIndicator <- but
				fmt.Println("3333")
			}
		}
	}
}

func checkFloorArrival() {
	curFloorIndicator := ElevGetFloorSensorSignal()
	if curFloorIndicator != -1 && lastFloorIndicator == -1 {
		ElevSetFloorIndicator(curFloorIndicator)
		//Sette eventAtFloor-channel
	}
	lastFloorIndicator = curFloorIndicator
}

func setDoor(doorOpen int) {
	ElevSetDoorOpenLamp(doorOpen)
}
