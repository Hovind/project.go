package order

import (
    "fmt"
    //"math"
    . "project.go/obj"
)

const (
    N_FLOORS   = 4
    travelTime = 1
    stopTime   = 3
)

const (
    UP = 0 + iota
    DOWN
    COMMAND
)

const (
    CONTINUE = 0 + iota
    STOP
    OPEN_DOOR
)

type cart struct {
    floor, direction int
    commands         [4]bool
}

type Orders struct {
    local_addr string
    hall       [4][2]bool
    carts      map[string]*cart
}

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

func order_manager(light_channel chan<- Order) (chan<- Order, chan<- int, chan chan int, chan chan int, chan chan int) {
    local_addr, to_network_channel, from_network_channel, _, _ := network.Manager("33223")

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
    order_request_channel := make(chan chan int);

    hall := [4][2]bool{};
    carts := make(map[string]cart);
    func add_order(order Order, addr string, hall [4][2]bool, carts map[string]cart) {
        if data.o.Button == COMMAND {
            if addr == local_addr {
                light_channel <-data.o;
            }
            carts[addr].commands[order.Floor] = order.Value;
        } else {
            light_channel <-data.o;
            hall[order.Floor][order.Button] = value;
        }
    }

    func map_key_exists(carts map[string]cart, addr string) {
        _, ok := carts[addr];
        return ok;
    }

    go func() {
        floor := -1;
        for {
            //system.Print();
            select {
            case data := <-order_from_network_channel:
                if !map_key_exists(carts, addr) {
                    carts[addr] = cart{};
                }
                add_order(data.o, data.a, hall, carts);
            case data := <-sync_from_network_channel:
                if data.a == local_addr {
                    hall = data.s.hall;
                    carts = data.s.carts;
                } else {
                    //sync(system
                    sync_to_network_channel <-data.s;
                }
            case data := <-floor_from_network_channel:
                if !map_key_exists(carts, addr) {
                    system.AddCartToMap(order.NewCart(), data.a)
                }
                carts[data.a].Floor = data.v;
            case data := <-direction_from_network_channel:
                if !system.CheckIfCart(data.a) {
                    carts[addr] = cart{};
                }
                carts[addr].Direction = data.v;
            case order := <-order_channel:
                add_order(order, local_addr, hall, carts);
                order_to_network_channel <-order;
            case floor = <-floor_channel:
                carts[local_addr].Floor = floor;
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
            case response_channel := <-order_request_channel:
                if new_order {
                    response_channel <-1;
                    new_order = false;
                } else {
                    response_channel <-0;
                }
            }
        }
    }()
    return order_channel, floor_channel, stop_request_channel, direction_request_channel, order_request_channel;
}


//SAVED FOR LATER
func NewOrders(local_addr string) *Orders {
    o := &Orders{local_addr: local_addr, carts: make(map[string]*cart)}
    o.carts[local_addr] = &cart{floor: -1, direction: -1}
    return o
}

func (self Orders) Print() {
    fmt.Println("Local address:", self.local_addr)
    fmt.Println("Hall\n", self.hall)
    for key, value := range self.carts {
        fmt.Println("IP:", key)
        fmt.Println("Cart:", value.commands)
        fmt.Println("Floor:", value.floor);
        fmt.Println("Direction:", value.direction, "\n");
    }
}
func NewCart() *cart {
    return &cart{}
}

func (self Orders) CheckIfCart(addr string) bool {
    _, ok := self.carts[addr]
    return ok
}

func (self *Orders) AddCartToMap(cart *cart, name string) {
    self.carts[name] = cart
}

func (self Orders) CurDir(addr string) int {
    return self.carts[addr].curDir()
}

func (self Orders) CurFloor(addr string) int {
    return self.carts[addr].curFloor()
}

func (self *Orders) SetDir(addr string, direction int) {
    self.carts[addr].setDir(direction)
}

func (self *Orders) SetFloor(addr string, floor int) {
    self.carts[addr].setFloor(floor)
}



func (self *Orders) SetHallOrder(floor int, button int, value bool) {
    self.hall[floor][button] = value
}

func (self Orders) checkIfHallOrder(floor, direction int) bool {
    if direction == -1 {
        return self.hall[floor][1]
    } else if direction == 1 {
        return self.hall[floor][0]
    } else {
        return  self.hall[floor][0] ||
                self.hall[floor][1]
    }

}

func (self *Orders) SetCommand(addr string, floor int, value bool) {
    self.carts[addr].commands[floor] = value
}

func (self cart) checkIfCommand(floor int) bool {
    return self.commands[floor]
}

func (self cart) curDir() int {
    return self.direction
}

func (self cart) curFloor() int {
    return self.floor
}

func (self *cart) setFloor(floor int) {
    self.floor = floor
}

func (self *cart) setDir(direction int) {
    self.direction = direction
}

func (self Orders) CheckIfStopOnFloor(orderFloor int, orderDirection int) bool {
    return self.carts[self.local_addr].checkIfCommand(orderFloor) ||
        self.checkIfHallOrder(orderFloor, orderDirection) ||
        !self.CheckIfOrdersInDirection(orderFloor, orderDirection)
}

func (self Orders) Alone() bool {
    fmt.Println("LENGTH:", self.carts)
    return len(self.carts) == 1;
}

func (self Orders) CheckFloorAction(orderFloor int, orderDirection int) int {
    fmt.Println("Floor action on floor", orderFloor, "and direction", orderDirection)
    if  self.carts[self.local_addr].checkIfCommand(orderFloor) ||
        self.checkIfHallOrder(orderFloor, orderDirection) {
        fmt.Println("Traitor lines.")
        return OPEN_DOOR;
    } else if !self.CheckIfOrdersInDirection(orderFloor, orderDirection) {
        return STOP;
    } else if !self.search_for_orders_in_direction(orderFloor, orderDirection) {
        fmt.Println("You what.");
        return OPEN_DOOR;
    } else {
        return CONTINUE;
    }
}

func check_floor_action(floor, direction int) {

}

func (self Orders) get_orders_in_direction(floor, direction int) []Order {
    orders := []Order{};
    if direction == 0 {
        return orders;
    }
    for f := floor + direction; f != -1 && f != N_FLOORS; f += direction {
        fmt.Println("Looking for order at floor", f, "and direction", direction);
        if self.checkIfHallOrder(f, 1) {
            orders = append(orders, Order{Button: UP, Floor: f, Value: true});
        }
        if self.checkIfHallOrder(f, -1) {
            orders = append(orders, Order{Button: DOWN, Floor: f, Value: true});
        }
    }
    return orders;
}
func (self Orders) search_for_orders_in_direction(floor, direction int) bool {
    orders_in_direction := self.get_orders_in_direction(floor, direction);
    fmt.Println("Orders in dir:", orders_in_direction);
    for _, o := range orders_in_direction {
        direction := 1;
        if o.Button == DOWN {
            direction = -1;
        }
        if self.orderIsBestForMe(o.Floor, direction) {
            return true;
        }
    }
    return false;
}


func (self Orders) CheckIfOrdersInDirection(floor, direction int) bool {
    if direction == 0 {
        return false
    }
    for f := floor + direction; f != N_FLOORS && f != -1; f += direction {
        if self.carts[self.local_addr].checkIfCommand(f) || self.checkIfHallOrder(f, -1) || self.checkIfHallOrder(f, 1) {
            return true
        }
    }
    return false
}


func (self Orders) addedCostForElevator(cart *cart, orderFloor int, orderDirection int) int {
    turnFloor := cart.curFloor()
    lastFloor := turnFloor
    if cart.curDir() == 0 {
        return abs(cart.curFloor()-orderFloor) * travelTime
    }
    for floor := cart.curFloor(); floor < N_FLOORS; floor += cart.curDir() {
        if floor < 0 {
            break
        } else if floor == orderFloor && cart.curDir() == orderDirection {
            return 0
        } else if cart.checkIfCommand(floor) {
            turnFloor = floor
        }
    }
    for floor := (turnFloor - cart.curDir()); floor < N_FLOORS; floor -= cart.curDir() {
        if floor < 0 {
            break
        } else if floor == orderFloor {
            return 0
        } else if cart.checkIfCommand(floor) {
            lastFloor = floor
        }
    }
    return abs(lastFloor-orderFloor) * travelTime
}

func (self Orders) costForClient(cart *cart, orderFloor int, orderDirection int) int {
    stops := 0
    turnFloor := cart.curFloor()
    noStops := true
    if cart.curDir() == 0 {
        return abs(orderFloor-cart.curFloor()) * travelTime
    }
    for floor := cart.curFloor(); floor < N_FLOORS; floor += cart.curDir() {
        if floor < 0 {
            break
        } else if floor == orderFloor && orderDirection == cart.curDir() || floor == 0 || floor == N_FLOORS -1 {
            return abs(cart.curFloor()-floor)*travelTime + stops*stopTime
        } else if cart.checkIfCommand(floor) {
            stops += 1
            turnFloor = floor
            noStops = false
        }
    }
    floor := turnFloor - 1
    for floor < N_FLOORS {
        if floor < 0 {
            if noStops {
                return abs(cart.curFloor() - orderFloor)*travelTime;
            }
            return (abs(cart.curFloor()-turnFloor)+abs(floor-orderFloor)+N_FLOORS-1)*travelTime + stops*stopTime
        } else if floor == orderFloor && orderDirection == -cart.curDir() || floor == 0 || floor == N_FLOORS -1 {
            return (abs(cart.curFloor()-turnFloor)+abs(floor-orderFloor))*travelTime + stops*stopTime
        } else if cart.checkIfCommand(floor) {
            stops += 1
            noStops = false
        }
        floor -= cart.curDir()
    }
    return 0
}


func (self cart) furthest_command(floor, direction int) int {
    if direction == /*elev.STOP*/0 {
        return floor;
    }
    f := 0;
    if direction == /*elev.UP*/1 {
        f = N_FLOORS - 1;
    }
    for f != floor {
        if self.checkIfCommand(f) {
            return f;
        }
        f -= direction;
    }
    return floor;
}

func abs(x int) int {
    if x < 0 {
        return -x
    } else {
        return x
    }
}

func sum(commands [N_FLOORS]bool) int {
    sum := 0;
    for _, e := range commands {
        if e {
            sum += 1;
        }
    }
    return sum;
}
func max(a, b int) int {
    if a > b {
        return a;
    } else {
        return b;
    }
}

func sign(a int) int {
    if a > 0 {
        return 1;
    } else if a < 0 {
        return -1;
    } else {
        return 0;
    }
}

func (self cart) cost(floor, direction int) int {
    turn_floor := self.curFloor();
    final_floor := floor;
    if self.curDir() == sign(floor - self.curFloor()) {
        turn_floor = self.curFloor() + self.curDir() * max(abs(self.furthest_command(self.curFloor(), self.curDir()) - self.curFloor()), abs(floor - self.curFloor()))
        final_floor = self.furthest_command(turn_floor, -self.curDir());
    } else if self.curDir() == -sign(floor - self.curFloor()) {
        turn_floor = self.furthest_command(self.curFloor(), self.curDir());
        final_floor = turn_floor - self.curDir() * max(abs(self.furthest_command(turn_floor, -self.curDir()) - turn_floor), abs(floor - turn_floor));
    }
    return sum(self.commands) * stopTime + (abs(turn_floor - self.curFloor()) + abs(final_floor - turn_floor)) * travelTime;
}




func (self Orders) orderIsBestForMe(orderFloor int, orderDirection int) bool {
    lowestCost := 300
    bestCart_addr := ""
    for addr, cart := range self.carts {
        //clientCost := self.costForClient(cart, orderFloor, orderDirection)
        addedCost := cart.cost(orderFloor, orderDirection) - cart.cost(cart.curFloor(), cart.curDir());//self.addedCostForElevator(cart, orderFloor, orderDirection)
        cost := /*clientCost*3*/ + addedCost*1
        fmt.Println("IP:", addr, "Cost:", cost);
        if cost < lowestCost || cost == lowestCost && addr < bestCart_addr {
            lowestCost = cost
            bestCart_addr = addr
        }
    }
    return bestCart_addr == self.local_addr
}

func (self Orders) GetDirection() int {
    curDir := self.carts[self.local_addr].curDir()
    curFloor := self.carts[self.local_addr].curFloor()
    iterateDirection := curDir
    if iterateDirection == 0 {
        iterateDirection = 1
    }
    floor := curFloor + iterateDirection;
    for floor < N_FLOORS {
        if floor < 0 {
            break
        } else if self.carts[self.local_addr].checkIfCommand(floor) {
            fmt.Println("Found command in dir")
            return iterateDirection
        } else if (self.checkIfHallOrder(floor, iterateDirection) &&
            (self.orderIsBestForMe(floor, iterateDirection))) ||
            (self.checkIfHallOrder(floor, -iterateDirection) &&
            (self.orderIsBestForMe(floor, -iterateDirection))) {
            fmt.Println("Found order for me in direction")
            return iterateDirection
        }
        floor += iterateDirection
    }
    floor = curFloor - iterateDirection
    for floor < N_FLOORS {
        if floor < 0 {
            break
        } else if self.carts[self.local_addr].checkIfCommand(floor) {
            fmt.Println("Command in -dir, not in dir")
            return -iterateDirection
        } else if (self.checkIfHallOrder(floor, iterateDirection) &&
            (self.orderIsBestForMe(floor, iterateDirection))) ||
            (self.checkIfHallOrder(floor, -iterateDirection) &&
                (self.orderIsBestForMe(floor, -iterateDirection))) {
            fmt.Println("Found order for me in -dir, not command or in dir.")
            return -iterateDirection
        }
        floor -= iterateDirection
    }
    fmt.Println("Found no orders.")
    return 0
}

func (self *Orders) Sync(s Orders) {
    for f := range s.hall {
        for b := range s.hall[f] {
            s.hall[f][b] = self.hall[f][b];
        }
    }
    for addr, cart := range s.carts {
        self.carts[addr] = cart;
    }
}
