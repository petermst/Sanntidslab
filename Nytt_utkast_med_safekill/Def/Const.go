package Def

const (
	STATE_IDLE      State = 0
	STATE_MOVING    State = 1
	STATE_DOOR_OPEN State = 2
	STATE_STUCK     State = 3
)

const (
	N_BUTTONS int = 3
	N_FLOORS      = 4
)

const (
	MOTOR_DOWN int = -1
	MOTOR_IDLE int = 0
	MOTOR_UP   int = 1
)

const (
	DOOR_CLOSE int = 0
	DOOR_OPEN      = 1
)

const (
	IS_STOPPED int = 1
	IS_MOVING      = 0
)
