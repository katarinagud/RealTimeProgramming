package fsm 

import "../elevio"
import "../types"
import "time"
import "fmt"
import "encoding/json"
import "io/ioutil"






func Fsm_run_elev(newOrder <-chan types.Button, floorReached <-chan int, orderDone chan types.Button, local_state chan types.ElevState) {	

	//Initializing ElevState

	e := types.ElevState{}
	{
		f := elevio.GetFloor()
		if f == -1 {
			e.Floor     = 0
			e.Direction = elevio.MD_Down
			e.State     = types.MOVING
			elevio.SetMotorDirection(elevio.MD_Down)
		} else {
			e.Floor     = f
			e.Direction = elevio.MD_Stop
			e.State     = types.IDLE
		}
	}
	
	doorTime := time.NewTimer(3*time.Second)
	doorTime.Stop()	
	
	for{
		select{
		case newOrder := <- newOrder:

			e.Orders[newOrder.Floor][newOrder.Type] = 1
			local_state <- e

			switch e.State {
			case types.IDLE:
				e.Direction = ChooseDirection(e)

				if (e.Direction == elevio.MD_Stop) && ShouldStop(e) {
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					doorTime.Reset(3*time.Second)										
					local_state <- e

				} else {
					elevio.SetMotorDirection(ChooseDirection(e))
					e.State = types.MOVING
					local_state <- e
				}				
				
			case types.MOVING:
				e.Direction = ChooseDirection(e)
				local_state <- e

			case types.DOOR_OPEN:
				if e.Floor == newOrder.Floor {
					e.State = types.DOOR_OPEN
					doorTime.Reset(3*time.Second)
					local_state <- e
				}				
			}
		
		case floorReached := <- floorReached:
			elevio.SetFloorIndicator(floorReached)
			e.Floor = floorReached
			local_state <- e

			switch e.State {			
			case types.MOVING:
				
				if ShouldStop(e) {
					elevio.SetMotorDirection(0)
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					e = ClearAtCurrentFloor(e, func(btn int){ orderDone <- types.Button{e.Floor, btn}})		
					doorTime.Reset(3*time.Second)
					local_state <- e
				}

			case types.IDLE:
				if ShouldStop(e) {
					elevio.SetMotorDirection(0)
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					e = ClearAtCurrentFloor(e, func(btn int){ orderDone <- types.Button{e.Floor, btn}})		
					doorTime.Reset(3*time.Second)
					local_state <- e
				}

			case types.INIT:
				elevio.SetMotorDirection(0)
				e.State = types.IDLE
				e.Floor = floorReached
				e.Direction = elevio.MD_Stop
				local_state <- e
			}

		case <- doorTime.C:
			
			switch e.State {
			case types.DOOR_OPEN:
				elevio.SetDoorOpenLamp(false)
				e.State = types.IDLE
				e = ClearAtCurrentFloor(e, func(btn int){ orderDone <- types.Button{e.Floor, btn}})		
				dir := ChooseDirection(e)
				elevio.SetMotorDirection(dir)
				e.Direction = dir
				
				if dir != elevio.MD_Stop {
					e.State = types.MOVING
				}
				local_state <- e
			}
		}
	}
}


//Writes cab orders to a json-file, works as a backup when an elevator looses power.

func WriteCabOrdersToFile(localStateToFsm <-chan types.ElevState, newOrder chan<- types.Button) {
	var cabOrders [4]bool
	str, _ := ioutil.ReadFile("cabOrderBackup.json")		
	fmt.Printf("Reading from file")
	json.Unmarshal(str, &cabOrders)
	for f := 0; f < 4; f++ {
		if cabOrders[f] {
			elevio.SetButtonLamp(2, f, true)
			newOrder <- types.Button{f, 2}
		}
	}

	for{
		fmt.Printf("in for-loop")
		state := <- localStateToFsm
		var cabOrders [4]bool
		for f := 0; f < 4; f++ {
			if state.Orders[f][2] != 0 {
				cabOrders[f] = true
			}
		}
		str, _ := json.Marshal(cabOrders)
		fmt.Printf("writeing to file")
		ioutil.WriteFile("cabOrderBackup.json", str, 0666)
	}
}	



