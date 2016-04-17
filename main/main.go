package main

import (
    "encoding/json"
    "fmt"
    "time"
    //"os"

    "project.go/elev"
    "project.go/network"
    . "project.go/obj"
    "project.go/order"
    "project.go/utils"
)

func network_decoder(from_network_channel <-chan Message) (<-chan struct{o Order; a string}, <-chan struct{s order.Orders; a string}, <-chan struct{v int; a string}, <-chan struct{v int; a string}) {
    order_from_network_channel := make(chan struct{o Order; a string});
    sync_from_network_channel := make(chan struct{s order.Orders; a string});
    floor_from_network_channel := make(chan struct{v int; a string});
    direction_from_network_channel := make(chan struct{v int; a string});
    go func() {
        for {
            msg := <-from_network_channel;
            addr := msg.Origin.IP.String();

            v, o, s, err := 0, Order{}, order.Orders{}, error(nil)
            switch msg.Code {
            case ORDER:
                err = json.Unmarshal(msg.Body, &o)
            case FLOOR_UPDATE, DIRECTION_UPDATE:
                err = json.Unmarshal(msg.Body, &v)
            case SYNC:
                err = json.Unmarshal(msg.Body, &s);
            }
            if err != nil {
                fmt.Println("Could not unmarshal order.")
                continue;
            }
            switch msg.Code {
            case ORDER:
                data := struct{o Order; a string}{o, addr}
                order_from_network_channel <-data;
            case FLOOR_UPDATE:
                data := struct{v int; a string}{v, addr}
                floor_from_network_channel <-data;
            case DIRECTION_UPDATE:
                data := struct{v int; a string}{v, addr}
                direction_from_network_channel <-data;
            case SYNC:
                data := struct{s order.Orders; a string}{s, addr};
                sync_from_network_channel <-data;
            }
        }
    }();
    return order_from_network_channel, sync_from_network_channel, floor_from_network_channel, direction_from_network_channel;
}

func network_encoder(to_network_channel chan<- Message) (chan<- Order, chan<- order.Orders, chan<- int, chan<- int) {
    order_to_network_channel := make(chan Order);
    sync_to_network_channel := make(chan order.Orders);
    floor_to_network_channel := make(chan int);
    direction_to_network_channel := make(chan int);
    go func() {
        for {
            select {
            case order := <- order_to_network_channel:
                b, err := json.Marshal(order);
                if err != nil {
                    fmt.Println("Could not marshal order.");
                } else {
                    to_network_channel <-*NewMessage(ORDER, b, nil, nil);
                }
            case orders := <-sync_to_network_channel:
                b, err := json.Marshal(orders);
                if err != nil {
                    fmt.Println("Could not marshal order.");
                } else {
                    to_network_channel <-*NewMessage(SYNC, b, nil, nil);
                }
            case floor := <-floor_to_network_channel:
                b, err := json.Marshal(floor);
                if err != nil {
                    fmt.Println("Could not marshal floor.");
                } else {
                    to_network_channel <-*NewMessage(FLOOR_UPDATE, b, nil, nil);
                }
            case direction := <-direction_to_network_channel:
                b, err := json.Marshal(direction);
                if err != nil {
                    fmt.Println("Could not marshal direction.");
                } else {
                    to_network_channel <-*NewMessage(DIRECTION_UPDATE, b, nil, nil);
                }
            }
        }
    }();
    return order_to_network_channel, sync_to_network_channel, floor_to_network_channel, direction_to_network_channel;
}

func order_manager(light_channel chan<- Order) (chan<- Order, chan<- int, chan chan int, chan chan int) {
    local_addr, to_network_channel, from_network_channel := network.Manager("33223")



    order_to_network_channel,
    sync_to_network_channel,
    floor_to_network_channel,
    direction_to_network_channel := network_encoder(to_network_channel);

    order_from_network_channel,
    sync_from_network_channel,
    floor_from_network_channel,
    direction_from_network_channel := network_decoder(from_network_channel);

    order_channel := make(chan Order);
    floor_channel := make(chan int);
    stop_request_channel := make(chan chan int);
    direction_request_channel := make(chan chan int);

    system := order.NewOrders(local_addr)
    go func() {
        floor := 0;
        new_order := false;
        for {
            select {
            case data := <-order_from_network_channel:
                if !system.CheckIfCart(data.a) {
                    system.AddCartToMap(order.NewCart(), data.a)
                }
                if data.o.Button == order.COMMAND {
                    if data.a == local_addr {
                        light_channel <-data.o;
                    }
                    system.SetCommand(data.a, data.o.Floor, data.o.Value)
                } else {
                    light_channel <-data.o;
                    system.SetHallOrder(data.o.Floor, data.o.Button, data.o.Value)
                    //hall[order.Floor][order.Button] = value;
                }
            case data := <-sync_from_network_channel:
                if data.s.Addr == local_addr {
                    system.Sync(&data.s, &new_order, light_channel);
                } else {
                    data.s.Sync(system, &new_order, light_channel);
                    sync_to_network_channel <-data.s;
                }
            case data := <-floor_from_network_channel:
                if !system.CheckIfCart(data.a) {
                    system.AddCartToMap(order.NewCart(), data.a)
                }
                system.SetFloor(data.a, data.v)
            case data := <-direction_from_network_channel:
                if !system.CheckIfCart(data.a) {
                    system.AddCartToMap(order.NewCart(), data.a)
                }
                system.SetDir(data.a, data.v)
            case order := <-order_channel:
                order_to_network_channel <-order;
            case floor = <-floor_channel:
                system.SetFloor(local_addr, floor);
                //carts[local_addr].Floor = floor;
                floor_to_network_channel <-floor;
            case response_channel := <-stop_request_channel:
                floor_action := system.CheckFloorAction(floor, system.CurDir(local_addr));
                if floor_action == order.OPEN_DOOR {
                    order_to_network_channel <-Order{Button: order.COMMAND, Floor: floor, Value: false}
                }
                response_channel <-floor_action;
            case response_channel := <-direction_request_channel:
                direction := system.GetDirection()
                system.SetDir(local_addr, direction)
                direction_to_network_channel <-direction;

                button := order.UP
                floor := system.CurFloor(local_addr)
                if direction == elev.DOWN {
                    button = order.DOWN
                } else if direction == elev.STOP {

                    order_to_network_channel <-Order{Button: order.DOWN, Floor: floor, Value: false}
                }
                order_to_network_channel <-Order{Button: button, Floor: floor, Value: false}
                //carts[local_addr].Direction = direction;
                response_channel <- direction
            }
        }
    }()
    return order_channel, floor_channel, stop_request_channel, direction_request_channel;
}

func light_manager() chan<- Order {
    light_channel := make(chan Order);

    go func() {
        for {
            order := <-light_channel;
            elev.SetButtonLamp(order.Button, order.Floor, order.Value);
        }
    }();
    return light_channel;
}

func main() {
    door_open := false;
    door_timer := utils.NewTimer();
    active_timer := utils.NewTimer();

    elev.Init();
    set_direction(elev.DOWN, active_timer);

    button_channel := elev.Button_checker();
    floor_sensor_channel := elev.Floor_checker();
    stop_button_channel := elev.Stop_checker();

    light_channel := light_manager();
    order_channel, floor_channel, stop_request_channel, direction_request_channel := order_manager(light_channel);

    floor := -1;
    direction := elev.DOWN;
    for {
        select {
        case order := <-button_channel:
            if floor == order.Floor && door_open {
                door_timer.Start(3*time.Second);
            } else {
                order_channel <-order;
            }
        case floor = <-floor_sensor_channel:
            elev.SetFloorIndicator(floor);
            floor_channel <-floor;
            floor_action := request(stop_request_channel);
            if floor_action == order.OPEN_DOOR {
                open_door(door_timer, active_timer, &door_open);
                direction = elev.STOP;
            } else if floor_action == order.STOP {
                stop(active_timer);
                direction = elev.STOP;
            }
        case <-stop_button_channel:
            stop(active_timer);
        case <-door_timer.Timer.C:
            close_door(&door_open);
            direction = request(direction_request_channel);
            set_direction(direction, active_timer);
        case <-active_timer.Timer.C:
            stop(active_timer);
            return;
        case <-time.After(500*time.Millisecond):
            floor_action := request(stop_request_channel);
            if !door_open && floor_action == order.OPEN_DOOR && direction == elev.STOP {
                open_door(door_timer, active_timer, &door_open);
            } else if !door_open && direction == elev.STOP {
                direction = request(direction_request_channel);
                set_direction(direction, active_timer);
            }
        }
    }
}

func stop(active_timer *utils.Timer) {
    elev.SetMotorDirection(elev.STOP);
}

func close_door(door_open *bool) {
    *door_open = false;
    elev.SetDoorOpenLamp(false);
}

func open_door(door_timer, active_timer *utils.Timer, door_open *bool) {
    *door_open = true;
    elev.SetDoorOpenLamp(true);
    elev.SetMotorDirection(elev.STOP);
    door_timer.Start(3*time.Second);
    active_timer.Stop();
}

func set_direction(direction int, active_timer *utils.Timer) {
    elev.SetMotorDirection(direction);
    active_timer.Start(10 * time.Second);
}

func request(request_channel chan chan int) int {
    response_channel := make(chan int);
    request_channel <-response_channel;
    value := <-response_channel;
    close(response_channel);
    return value;
}
