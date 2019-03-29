package types

import "../elevio"

type Button struct {
	Floor int
	Type int
}

const (
	IDLE = iota
	MOVING
	DOOR_OPEN
	INIT
)

type ElevState struct {
	Floor int
	Direction elevio.MotorDirection
	State int
	Orders [4][3] int
}

type Order struct {
	Floor int
	Button elevio.ButtonType
	AssignedTo string
}

const N_BUTTONS = 3
const N_FLOORS = 4