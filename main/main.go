package main

import (
	"fmt"
	"os"
	"project.go/elev"
	. "project.go/obj"
	"project.go/ord"
	"project.go/utils"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide port.")
		return
	}
	port := os.Args[1]
	door_timer := utils.New_timer()
	active_timer := utils.New_timer()

	elev.Init()
	set_direction(Direction.Down, active_timer)

	button_channel := elev.Button_checker()
	floor_sensor_channel := elev.Floor_checker()
	stop_button_channel := elev.Stop_checker()
	light_channel := elev.Light_manager()

	order_channel, floor_channel, stop_request_channel, direction_request_channel := ord.Manager(port, light_channel)

	door_open := false
	floor := 0
	direction := Direction.Down
	for {
		select {
		case order := <-button_channel:
			if floor == order.Floor && door_open {
				door_timer.Start(3 * time.Second)
			} else {
				order_channel <- order
			}
		case floor = <-floor_sensor_channel:
			elev.Set_floor_indicator(floor)
			floor_channel <- floor
			floor_action := request(stop_request_channel)
			if floor_action == ord.OPEN_DOOR {
				open_door(door_timer, active_timer, &door_open)
				direction = Direction.Stop
			} else if floor_action == ord.STOP {
				stop(active_timer)
				direction = Direction.Stop
			}
		case <-stop_button_channel:
			stop(active_timer)
		case <-door_timer.Timer.C:
			close_door(&door_open)
			direction = request(direction_request_channel)
			set_direction(direction, active_timer)
		case <-active_timer.Timer.C:
			stop(active_timer)
			fmt.Println("Please refrain from sabotaging the elevator. Thank you.")
			return
		case <-time.After(500 * time.Millisecond):
			floor_action := request(stop_request_channel)
			if !door_open && floor_action == ord.OPEN_DOOR && direction == Direction.Stop {
				open_door(door_timer, active_timer, &door_open)
			} else if !door_open && direction == Direction.Stop {
				direction = request(direction_request_channel)
				set_direction(direction, active_timer)
			}
		}
	}
}

func stop(active_timer *utils.Timer) {
	elev.Set_motor_direction(Direction.Stop)
	active_timer.Stop()
}

func close_door(door_open *bool) {
	*door_open = false
	elev.Set_door_open_lamp(false)
}

func open_door(door_timer, active_timer *utils.Timer, door_open *bool) {
	*door_open = true
	elev.Set_door_open_lamp(true)
	elev.Set_motor_direction(Direction.Stop)
	door_timer.Start(3 * time.Second)
	active_timer.Stop()
}

func set_direction(direction int, active_timer *utils.Timer) {
	elev.Set_motor_direction(direction)
	active_timer.Start(10 * time.Second)
}

func request(request_channel chan chan int) int {
	response_channel := make(chan int)
	request_channel <- response_channel
	value := <-response_channel
	close(response_channel)
	return value
}
