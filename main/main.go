package main

import (
	"encoding/json"
	"fmt"
	"time"

	"project.go/elev"
	"project.go/network"
	. "project.go/obj"
	"project.go/order"
	"project.go/timer"
)

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

func Manager(push_light_channel, pop_light_channel chan<- Order, open_door_channel, direction_channel chan<- int) (chan<- Order, chan<- int, chan<- int) {
	local_addr, to_network_channel, from_network_channel, _, _ := network.Manager("33223")

	order_channel := make(chan Order)
	check_stop_channel := make(chan int)
	direction_request_channel := make(chan int)
	system := order.NewOrders(local_addr)
	go func() {
		for {

			//system.Print();
			select {
			case msg := <-from_network_channel:
				addr := msg.Origin.IP.String()
				if !system.CheckIfCart(addr) {
					system.AddCartToMap(order.NewCart(), addr)
				}
				system.Print()
				err := (error)(nil)
				o := Order{}
				i := 0
				switch msg.Code {
				case ORDER_PUSH, ORDER_POP:
					err = json.Unmarshal(msg.Body, &o)
				case FLOOR_HIT, DIRECTION_CHANGE:
					err = json.Unmarshal(msg.Body, &i)
				}
				if err != nil {
					fmt.Println("Could not unmarshal order.")
					break
				}
				switch msg.Code {
				case ORDER_PUSH, ORDER_POP:
					value := true
					if msg.Code == ORDER_PUSH {
						push_light_channel <- o
						if system.CurDir(local_addr) == 0 { //  carts[local_addr].Direction == 0 {
							direction_channel <- system.GetDirection()
						}
					} else {
						value = false
						pop_light_channel <- o
					}
					fmt.Println("Setting, floor:", o.Floor, "button:", o.Button, value)
					if o.Button == order.COMMAND {
						system.SetCommand(addr, o.Floor, value)
					} else {
						system.SetHallOrder(o.Floor, o.Button, value)

						//hall[order.Floor][order.Button] = value;
					}
				case FLOOR_HIT:
					system.SetFloor(addr, i)
					//carts[addr].Floor = i;
				case DIRECTION_CHANGE:
					system.SetDir(addr, i)
					//carts[addr].Direction = i;
				}
			case order := <-order_channel:
				b, err := json.Marshal(order)
				if err != nil {
					fmt.Println("Could not marshal order.")
				} else {
					to_network_channel <- *NewMessage(ORDER_PUSH, b, nil, nil)
				}
			case floor := <-check_stop_channel:
				b, err := json.Marshal(floor)
				if err != nil {
					fmt.Println("Could not marshal direction.")
				} else {
					to_network_channel <- *NewMessage(FLOOR_HIT, b, nil, nil)
				}
				system.SetFloor(local_addr, floor)
				//carts[local_addr].Floor = floor;
				if system.CheckIfStopOnFloor(floor, system.CurDir(local_addr)) {
					o := Order{Button: order.COMMAND, Floor: floor}
					b, _ := json.Marshal(o)
					to_network_channel <- *NewMessage(ORDER_POP, b, nil, nil)
					fmt.Println("Sending pop, floor:", o.Floor, "button:", o.Button)
					open_door_channel <- floor
				}
			case <-direction_request_channel:
				direction := system.GetDirection()
				b, _ := json.Marshal(direction)
				to_network_channel <- *NewMessage(DIRECTION_CHANGE, b, nil, nil)
				button := order.UP
				floor := system.CurFloor()
				if direction == elev.DOWN {
					button = order.DOWN
				}
				o := Order{Button: button, Floor: floor}
				b, _ = json.Marshal(o)
				to_network_channel <- *NewMessage(ORDER_POP, b, nil, nil)
				fmt.Println("Sending pop, floor:", o.Floor, "button:", o.Button)
				system.SetDir(local_addr, direction)
				//carts[local_addr].Direction = direction;
				direction_channel <- direction
			}
		}
	}()
	return order_channel, check_stop_channel, direction_request_channel
}

func main() {
	elev.Init()
	door_open_floor := -1
	door_open := false
	door_timer := timer.New()
	elev.SetMotorDirection(-1)

	button_channel := elev.Button_checker()
	floor_signal_channel := elev.Floor_checker()
	stop_signal_channel := elev.Stop_checker()

	push_light_channel := make(chan Order)
	pop_light_channel := make(chan Order)
	open_door_channel := make(chan int)
	direction_channel := make(chan int)
	order_channel, check_stop_channel, direction_request_channel := Manager(push_light_channel, pop_light_channel, open_door_channel, direction_channel)

	for {
		select {
		case order := <-push_light_channel:
			elev.SetButtonLamp(order.Button, order.Floor, true)
		case order := <-pop_light_channel:
			elev.SetButtonLamp(order.Button, order.Floor, false)
		case door_open_floor = <-open_door_channel:
			door_open = true
			door_timer.Start(3 * time.Second)
			elev.SetDoorOpenLamp(true)
			elev.SetMotorDirection(0)
		case direction := <-direction_channel:
			if !door_open {
				elev.SetMotorDirection(direction)
			}
		case floor := <-floor_signal_channel:
			elev.SetFloorIndicator(floor)
			check_stop_channel <- floor
		case order := <-button_channel:
			if door_open && door_open_floor == order.Floor {
				door_timer.Start(3 * time.Second)
			} else {
				order_channel <- order
			}
		case <-door_timer.Timer.C:
			fmt.Println("DOOR TIMEOUT.")
			door_open = false
			elev.SetDoorOpenLamp(false)
			direction_request_channel <- door_open_floor
		case <-stop_signal_channel:
			elev.SetMotorDirection(0)
		}
	}
}
