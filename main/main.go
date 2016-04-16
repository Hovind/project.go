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
    "project.go/timer"
)

const (
    N_FLOORS  = 4
    N_BUTTONS = 3
)

func main() {
    door_open := false;
    door_timer := timer.New();

    elev.Init();
    elev.SetMotorDirection(elev.DOWN);

    button_channel := elev.Button_checker();
    floor_sensor_channel := elev.Floor_checker();
    stop_button_channel := elev.Stop_checker();
    light_channel := elev.light_manager();

    order_channel, floor_channel, stop_request_channel, direction_request_channel, order_request_channel := order_manager(light_channel);

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
                open_door(door_timer, &door_open);
            } else if floor_action == order.STOP {
                elev.SetMotorDirection(elev.STOP);
            }
            direction = elev.STOP;
        case <-stop_button_channel:
            elev.SetMotorDirection(elev.STOP);
        case <-door_timer.Timer.C:
            door_open = false;
            elev.SetDoorOpenLamp(false);
            direction = request(direction_request_channel);
            elev.SetMotorDirection(direction);
        case <-time.After(500*time.Millisecond):
            new_order := request(order_request_channel);
            if new_order == 1 {
                floor_action := request(stop_request_channel);
                if floor_action == order.OPEN_DOOR && direction == elev.STOP {
                    open_door(door_timer, &door_open);
                } else if !door_open && direction == elev.STOP {
                    direction = request(direction_request_channel);
                    elev.SetMotorDirection(direction);
                }
            }
        }
    }
}

func open_door(door_timer *timer.Timer, door_open *bool) {
    *door_open = true;
    elev.SetMotorDirection(elev.STOP);
    door_timer.Start(3*time.Second);
    elev.SetDoorOpenLamp(true);
}

func request(request_channel chan chan int) int {
    response_channel := make(chan int);
    request_channel <-response_channel;
    value := <-response_channel;
    close(response_channel);
    return value;
}
