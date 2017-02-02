package Driver

/*
#cgo CFLAGS: -std=gnu11
#cgo LDFLAGS: -lcomedi -lm
#include<stdio.h>
#include<stdlib.h>
#include "elev.h"
#include "channels.h"
#include "io.h"
*/
import "C"

func ElevInit() {
	C.elev_init()
}

func ElevSetMotorDirection(dirn int) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(dirn))
}

func ElevSetButtonLamp(button int, floor int, value int) {
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}

func ElevSetFloorIndicator(floor int) {
	C.elev_set_floor_indicator(C.int(floor))
}

func ElevSetDoorOpenLamp(value int) {
	C.elev_set_door_open_lamp(C.int(value))
}

func ElevSetStopLamp(value int) {
	C.elev_set_stop_lamp(C.int(value))
}

func ElevGetButtonSignal(button int, floor int) int {
	return int(C.elev_get_button_signal(C.elev_button_type_t(button), C.int(floor)))
}

func ElevGetFloorSensorSignal() int {
	return int(C.elev_get_floor_sensor_signal())
}

func ElevGetStopSignal() int {
	return int(C.elev_get_stop_signal())
}

func ElevGetObstructionSignal() int {
	return int(C.elev_get_obstruction_signal())
}
