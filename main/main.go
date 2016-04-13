package main
import (
    "project.go/elev"
    "project.go/network"
    "time"
    "fmt"
    "encoding/json"
    "project.go/order"
    . "project.go/obj"
)
/*
const (
    INIT = iota
    IDLE
    RUN
    DOOR_OPEN
)*/

const (
    N_FLOORS = 4
    N_BUTTONS = 3
    DOOR_TIMEOUT = 3
)

func Manager(push_light_channel, pop_light_channel chan<- Order, open_door_channel, direction_channel, check_stop_channel chan<- int) (chan<- Order, chan<- int, chan<- int) {
    local_addr, to_network_channel, from_network_channel, _, _ := network.Manager("33223");

    order_channel := make(chan Order);
    check_stop_channel := make(chan int);
    direction_request_channel := make(chan int);

    go func() {
        for {
            //system.Print();
            select {
            case msg := <-from_network_channel:
                addr := msg.Origin.IP.String();
                err := (error)(nil);
                order := &elev.Order{};
                i := (*int)(nil);
                switch msg.Code {
                case ORDER_PUSH, ORDER_POP:
                    err = json.Unmarshal(msg.Body, order);
                case FLOOR_HIT, DIRECTION_CHANGE:
                    err = json.Unmarshal(msg.Body, i);
                }
                if err != nil {
                    fmt.Println("Could not unmarshal order.");
                    break;
                }
                switch msg.Code {
                case ORDER_PUSH, ORDER_POP:
                    if msg.Code == ORDER_PUSH {
                        value := true;
                        push_light_channel <-order;
                        if carts[local_addr].Direction == 0 {
                            direction_channel <-getDirection();
                        }
                    } else {
                        value := false;
                        pop_light_channel <-order;
                    }
                    if order.Button == order.COMMAND {
                        cart[addr].Cab[order.Floor] = value;
                    } else {
                        hall[order.Floor][order.Button] = value;
                    }
                case FLOOR_HIT:
                    carts[addr].Floor = i;
                case DIRECTION_CHANGE:
                    carts[addr].Direction = i;
                }
            case order := <-order_channel:
                b, err := json.Marshal(order);
                if err != nil {
                    fmt.Println("Could not marshal order.");
                } else {
                    to_network_channel <-*NewMessage(ORDER_PUSH, b, nil, nil);
                }
            case floor := <-check_stop_channel:
                b, err := json.Marshal(floor);
                if err != nil {
                    fmt.Println("Could not marshal direction.");
                } else {
                    to_network_channel <-*NewMessage(FLOOR_HIT, b, nil, nil);
                }
                carts[local_addr].Floor = floor;
                if checkIfStop() {
                    order := Order{Button: order.COMMAND, Floor: floor};
                    b, _ := json.Marshal(order);
                    to_network_channel <-*NewMessage(ORDER_POP, b, nil, nil);
                    direction := getDirection();
                    if direction == elev.UP {
                        order.Button = order.UP;
                    } else if direction == elev.DOWN {
                        order.Button = order.DOWN;
                    }
                    b, _ := json.Marshal(order);
                    to_network_channel <-*NewMessage(ORDER_POP, b, nil, nil);
                    open_door_channel <-floor;
                }
            case <-direction_request_channel:
                direction := getDirection();
                b, err := json.Marshal(direction);
                if err != nil {
                    fmt.Println ("Could not marshal direction.");
                } else {
                    to_network_channel <-*NewMessage(DIRECTION_CHANGE, b, nil, nil);
                }
                carts[local_addr].Direction = direction;
                direction_channel <-direction;
            }
        }
    }();
    return order_channel, check_stop_channel, direction_request_channel;
}

func main() {
    elev.Init()
    door_open_floor := -1;
    door_open := false;

    button_channel := elev.Button_checker();
    floor_signal_channel := elev.Floor_checker();
    stop_signal_channel := elev.Stop_checker();

    push_light_channel := make(chan Order);
    pop_light_channel := make(chan Order);
    open_door_channel := make(chan int);
    direction_channel := make(chan int);

    order_channel, direction_request_channel := Manager(push_light_channel, pop_light_channel, open_door_channel, direction_channel);

    for {
        select {
        case order := <-push_light_channel:
            elev.SetButtonLamp(order.Floor, order.Button, true);
        case order := <-pop_light_channel:
            elev.SetButtonLamp(order.Floor, order.Button, false);
        case floor := <-open_door_channel:
            door_open = true;
            elev.SetDoorOpenLamp(true);
            elev.SetMotorDirection(0);
        case direction := <-direction_channel:
            elev.SetMotorDirection(direction);
        case floor := <-floor_signal_channel:
            check_stop_channel <-floor;
        case order := <-button_channel:
            if door_open && door_open_floor == order.Floor {
                door_timer.Start(3*time.Second);
            } else {
                order_channel <-order;
            }
        case <-door_timer.Timer.C:
            door_open = false;
            elev.SetDoorOpenLamp(false);
            direction_request_channel <-true;
        }
    }
}
