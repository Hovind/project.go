package ord

import (
    . "project.go/obj"
    "project.go/utils"
)

const (
    CONTINUE = 0 + iota
    STOP
    OPEN_DOOR
)

func check_if_hall_order(floor, direction int, hall[N_FLOORS][N_DIRECTIONS]bool) bool {
    if direction == Direction.Down {
        return hall[floor][Button.Down]
    } else if direction == Direction.Up {
        return hall[floor][Button.Up]
    } else {
        return  hall[floor][Button.Up] || hall[floor][Button.Down]
    }
}

func get_floor_action(cart_map map[string]*Cart, hall[N_FLOORS][N_DIRECTIONS]bool, local_addr string) int {
    cart := cart_map[local_addr];
    floor := cart.Floor;
    direction := cart.Direction;
    if  cart.Commands[floor] ||
        check_if_hall_order(floor, direction, hall) ||
        check_if_hall_order(floor, -direction, hall) &&
        !search_for_orders_in_direction(floor, direction, cart_map, hall, local_addr) &&
        furthest_command(floor, direction, cart.Commands) == floor {
        return OPEN_DOOR;
    } else if !search_for_orders_in_direction(floor, direction, cart_map, hall, local_addr) && !search_for_commands_in_direction(floor, direction, cart.Commands) {
        return STOP;
    } else {
        return CONTINUE;
    }
}

func get_orders_in_direction(floor, direction int, hall [N_FLOORS][N_DIRECTIONS]bool) []Order {
    orders := []Order{};
    if direction == 0 {
        return orders;
    }
    for f := floor + direction; f != -1 && f != N_FLOORS; f += direction {
        if hall[f][Button.Up] {
            orders = append(orders, Order{Button: Button.Up, Floor: f, Value: true});
        }
        if hall[f][Button.Down] {
            orders = append(orders, Order{Button: Button.Down, Floor: f, Value: true});
        }
    }
    return orders;
}

func search_for_orders_in_direction(floor, direction int, cart_map map[string]*Cart, hall [N_FLOORS][N_DIRECTIONS]bool, local_addr string) bool {
    orders_in_direction := get_orders_in_direction(floor, direction, hall);
    for _, o := range orders_in_direction {
        direction := Direction.Up;
        if o.Button == Button.Down {
            direction = Direction.Down;
        }
        if order_is_best_for_me(o.Floor, direction, cart_map, local_addr) {
            return true;
        }
    }
    return false;
}

func furthest_command(floor, direction int, commands [N_FLOORS]bool) int {
    if direction == Direction.Stop {
        return floor;
    }
    f := 0;
    if direction == Direction.Up {
        f = N_FLOORS - 1;
    }
    for f != floor {
        if commands[f] {
            return f;
        }
        f -= direction;
    }
    return floor;
}

func search_for_commands_in_direction(floor, direction int, commands [N_FLOORS]bool) bool {
    return furthest_command(floor, direction, commands) != floor;
}

func cost(floor, direction int, cart Cart) int {
    turn_floor := cart.Floor;
    final_floor := floor;
    if cart.Direction == utils.Sign(floor - cart.Floor) {
        turn_floor = cart.Floor + cart.Direction * utils.Max(utils.Abs(furthest_command(cart.Floor, cart.Direction, cart.Commands) - cart.Floor), utils.Abs(floor - cart.Floor))
        final_floor = furthest_command(turn_floor, -cart.Direction, cart.Commands);
    } else if cart.Direction == -utils.Sign(floor - cart.Floor) {
        turn_floor = furthest_command(cart.Floor, cart.Direction, cart.Commands);
        final_floor = turn_floor - cart.Direction * utils.Max(utils.Abs(furthest_command(turn_floor, -cart.Direction, cart.Commands) - turn_floor), utils.Abs(floor - turn_floor));
    }
    return utils.Sum(cart.Commands) + (utils.Abs(turn_floor - cart.Floor) + utils.Abs(final_floor - turn_floor));
}

func order_is_best_for_me(floor, direction int, cart_map map[string]*Cart, local_addr string) bool {
    min_weighted_cost := 0
    best_cart_addr := ""
    for addr, cart := range cart_map {
        previous_cost := cost(cart.Floor, cart.Direction, *cart);
        cost := cost(floor, direction, *cart)
        cost_difference :=  cost - previous_cost;
        weighted_cost := cost_difference;
        if best_cart_addr == "" || weighted_cost < min_weighted_cost || weighted_cost == min_weighted_cost && addr < best_cart_addr {
            min_weighted_cost = cost
            best_cart_addr = addr
        }
    }
    return best_cart_addr == local_addr
}

func get_direction(cart_map map[string]*Cart, hall [N_FLOORS][N_DIRECTIONS]bool, local_addr string) int {
    cart := cart_map[local_addr];
    floor := cart.Floor;
    direction := cart.Direction;
    if direction == 0 {
        direction = 1
    }
    if  search_for_orders_in_direction(floor, direction, cart_map, hall, local_addr) || search_for_commands_in_direction(floor, direction, cart_map[local_addr].Commands) {
        return direction;
    } else if search_for_orders_in_direction(floor, -direction, cart_map,  hall, local_addr) || search_for_commands_in_direction(floor, -direction, cart_map[local_addr].Commands) {
        return -direction;
    } else {
        return 0;
    }
}

func sync(s *Orders, hall *[N_FLOORS][N_DIRECTIONS]bool, cart_map map[string]*Cart, local_addr string, light_channel chan<- Order) {
    for f := range s.Hall {
        for b := range s.Hall[f] {
            value := s.Hall[f][b] || hall[f][b];
            if value {
                light_channel <-Order{Button: b, Floor: f, Value: true};
            }
            hall[f][b] = value;
            s.Hall[f][b] = value;
        }
    }
    c := cart_map[local_addr]
    for addr, _ := range cart_map {
        delete(cart_map, addr)
    }
    for addr, cart := range s.Carts {
        cart_map[addr] = cart;
    }
    cart_map[local_addr] = c;
    s.Carts[local_addr] = c;
}
