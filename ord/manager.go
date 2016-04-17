package ord

import (
	"encoding/json"
	"fmt"
	"project.go/network"
	. "project.go/obj"
)

func network_decoder(from_network_channel <-chan Message) (<-chan struct {
	o Order
	a string
}, <-chan struct {
	s Orders
	a string
}, <-chan struct {
	v int
	a string
}, <-chan struct {
	v int
	a string
}) {
	order_from_network_channel := make(chan struct {
		o Order
		a string
	})
	sync_from_network_channel := make(chan struct {
		s Orders
		a string
	})
	floor_from_network_channel := make(chan struct {
		v int
		a string
	})
	direction_from_network_channel := make(chan struct {
		v int
		a string
	})
	go func() {
		for {
			msg := <-from_network_channel
			addr := msg.Origin.IP.String()
			v, o, s, err := 0, Order{}, Orders{}, error(nil)
			switch msg.Code {
			case ORDER:
				err = json.Unmarshal(msg.Body, &o)
			case FLOOR_UPDATE, DIRECTION_UPDATE:
				err = json.Unmarshal(msg.Body, &v)
			case SYNC:
				err = json.Unmarshal(msg.Body, &s)
			}
			if err != nil {
				fmt.Println("Could not unmarshal order.")
				continue
			}
			switch msg.Code {
			case ORDER:
				data := struct {
					o Order
					a string
				}{o, addr}
				order_from_network_channel <- data
			case FLOOR_UPDATE:
				data := struct {
					v int
					a string
				}{v, addr}
				floor_from_network_channel <- data
			case DIRECTION_UPDATE:
				data := struct {
					v int
					a string
				}{v, addr}
				direction_from_network_channel <- data
			case SYNC:
				data := struct {
					s Orders
					a string
				}{s, addr}
				sync_from_network_channel <- data
			}
		}
	}()
	return order_from_network_channel, sync_from_network_channel, floor_from_network_channel, direction_from_network_channel
}

func network_encoder(to_network_channel chan<- Message) (chan<- Order, chan<- Orders, chan<- int, chan<- int) {
	order_to_network_channel := make(chan Order)
	sync_to_network_channel := make(chan Orders)
	floor_to_network_channel := make(chan int)
	direction_to_network_channel := make(chan int)
	go func() {
		for {
			select {
			case order := <-order_to_network_channel:
				b, err := json.Marshal(order)
				if err != nil {
					fmt.Println("Could not marshal order.")
				} else {
					to_network_channel <- *New_message(ORDER, b, nil, nil)
				}
			case orders := <-sync_to_network_channel:
				b, err := json.Marshal(orders)
				if err != nil {
					fmt.Println("Could not marshal order.")
				} else {
					to_network_channel <- *New_message(SYNC, b, nil, nil)
				}
			case floor := <-floor_to_network_channel:
				b, err := json.Marshal(floor)
				if err != nil {
					fmt.Println("Could not marshal floor.")
				} else {
					to_network_channel <- *New_message(FLOOR_UPDATE, b, nil, nil)
				}
			case direction := <-direction_to_network_channel:
				b, err := json.Marshal(direction)
				if err != nil {
					fmt.Println("Could not marshal direction.")
				} else {
					to_network_channel <- *New_message(DIRECTION_UPDATE, b, nil, nil)
				}
			}
		}
	}()
	return order_to_network_channel, sync_to_network_channel, floor_to_network_channel, direction_to_network_channel
}

func Manager(port string, light_channel chan<- Order) (chan<- Order, chan<- int, chan chan int, chan chan int) {
	local_addr, to_network_channel, from_network_channel := network.Manager(port)

	order_to_network_channel,
		sync_to_network_channel,
		floor_to_network_channel,
		direction_to_network_channel := network_encoder(to_network_channel)

	order_from_network_channel,
		sync_from_network_channel,
		floor_from_network_channel,
		direction_from_network_channel := network_decoder(from_network_channel)

	order_channel := make(chan Order)
	floor_channel := make(chan int)
	stop_request_channel := make(chan chan int)
	direction_request_channel := make(chan chan int)

	cart_map := make(map[string]*Cart)
	cart_map[local_addr] = New_cart()
	hall := [N_FLOORS][N_DIRECTIONS]bool{}
	go func() {
		floor := 0
		for {
			select {
			case data := <-order_from_network_channel:
				if _, ok := cart_map[data.a]; !ok {
					cart_map[data.a] = New_cart()
				}
				if data.o.Button == Button.Command {
					if data.a == local_addr {
						light_channel <- data.o
					}
					cart_map[data.a].Commands[data.o.Floor] = data.o.Value
				} else {
					light_channel <- data.o
					hall[data.o.Floor][data.o.Button] = data.o.Value
				}
			case data := <-sync_from_network_channel:
				if data.s.Addr == local_addr {
					sync(&data.s, &hall, cart_map, local_addr, light_channel)
				} else {
					sync(New_orders(local_addr, hall, cart_map), &data.s.Hall, data.s.Carts, data.s.Addr, light_channel)
					sync_to_network_channel <- data.s
				}
			case data := <-floor_from_network_channel:
				if _, ok := cart_map[data.a]; !ok {
					cart_map[data.a] = New_cart()
				}
				cart_map[data.a].Floor = data.v
			case data := <-direction_from_network_channel:
				if _, ok := cart_map[data.a]; !ok {
					cart_map[data.a] = New_cart()
				}
				cart_map[data.a].Direction = data.v
			case order := <-order_channel:
				order_to_network_channel <- order
			case floor = <-floor_channel:
				cart_map[local_addr].Floor = floor
				floor_to_network_channel <- floor
			case response_channel := <-stop_request_channel:
				floor_action := get_floor_action(cart_map, hall, local_addr)
				if floor_action == Action.Open_door {
					order_to_network_channel <- Order{Button: Button.Command, Floor: floor, Value: false}
				}
				response_channel <- floor_action
			case response_channel := <-direction_request_channel:
				direction := get_direction(cart_map, hall, local_addr)
				cart_map[local_addr].Direction = direction
				direction_to_network_channel <- direction

				button := Button.Up
				//floor := cart_map[local_addr].Floor
				if direction == Direction.Down {
					button = Button.Down
				} else if direction == Direction.Stop {
					order_to_network_channel <- Order{Button: Button.Down, Floor: floor, Value: false}
				}
				order_to_network_channel <- Order{Button: button, Floor: floor, Value: false}
				response_channel <- direction
			}
		}
	}()
	return order_channel, floor_channel, stop_request_channel, direction_request_channel
}
