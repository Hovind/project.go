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
    Floor, Direction int
    Commands         [4]bool
}

type Orders struct {
    Addr string
    Hall       [4][2]bool
    Carts      map[string]*cart
}


//SAVED FOR LATER
func NewOrders(local_addr string) *Orders {
    o := &Orders{Addr: local_addr, Carts: make(map[string]*cart)}
    o.Carts[local_addr] = &cart{Floor: -1, Direction: -1}
    return o
}


func (self Orders) Print() {
    fmt.Println("Local address:", self.Addr)
    fmt.Println("Hall\n", self.Hall)
    for key, value := range self.Carts {
        fmt.Println("IP:", key)
        fmt.Println("Cart:", value.Commands)
        fmt.Println("Floor:", value.Floor);
        fmt.Println("Direction:", value.Direction, "\n");
    }
}
func NewCart() *cart {
    return &cart{}
}

func (self Orders) CheckIfCart(addr string) bool {
    _, ok := self.Carts[addr]
    return ok
}

func (self *Orders) AddCartToMap(cart *cart, name string) {
    self.Carts[name] = cart
}

func (self Orders) CurDir(addr string) int {
    return self.Carts[addr].curDir()
}

func (self Orders) CurFloor(addr string) int {
    return self.Carts[addr].curFloor()
}

func (self *Orders) SetDir(addr string, direction int) {
    self.Carts[addr].setDir(direction)
}

func (self *Orders) SetFloor(addr string, floor int) {
    self.Carts[addr].setFloor(floor)
}



func (self *Orders) SetHallOrder(floor int, button int, value bool) {
    self.Hall[floor][button] = value
}

func (self Orders) checkIfHallOrder(floor, direction int) bool {
    if direction == -1 {
        return self.Hall[floor][1]
    } else if direction == 1 {
        return self.Hall[floor][0]
    } else {
        return  self.Hall[floor][0] ||
                self.Hall[floor][1]
    }

}

func (self *Orders) SetCommand(addr string, floor int, value bool) {
    self.Carts[addr].Commands[floor] = value
}

func (self cart) checkIfCommand(floor int) bool {
    return self.Commands[floor]
}

func (self cart) curDir() int {
    return self.Direction
}

func (self cart) curFloor() int {
    return self.Floor
}

func (self *cart) setFloor(floor int) {
    self.Floor = floor
}

func (self *cart) setDir(direction int) {
    self.Direction = direction
}
/*
func (self Orders) CheckIfStopOnFloor(orderFloor int, orderDirection int) bool {
    return self.Carts[self.Addr].checkIfCommand(orderFloor) ||
        self.checkIfHallOrder(orderFloor, orderDirection) ||
        !self.CheckIfOrdersInDirection(orderFloor, orderDirection)
}

func (self Orders) Alone() bool {
    fmt.Println("LENGTH:", self.Carts)
    return len(self.Carts) == 1;
}*/

func (self Orders) CheckFloorAction(orderFloor int, orderDirection int) int {
    fmt.Println("Floor action on floor", orderFloor, "and direction", orderDirection)
    if  self.Carts[self.Addr].checkIfCommand(orderFloor) ||
        self.checkIfHallOrder(orderFloor, orderDirection) ||
        self.checkIfHallOrder(orderFloor, -orderDirection) &&
        !self.search_for_orders_in_direction(orderFloor, orderDirection) &&
        self.Carts[self.Addr].furthest_command(orderFloor, orderDirection) == orderFloor {
        return OPEN_DOOR;
    } else if !self.search_for_orders_in_direction(orderFloor, orderDirection) {
        return STOP;
    } else {
        return CONTINUE;
    }
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

/*
func (self Orders) CheckIfOrdersInDirection(floor, direction int) bool {
    if direction == 0 {
        return false
    }
    for f := floor + direction; f != N_FLOORS && f != -1; f += direction {
        if self.Carts[self.Addr].checkIfCommand(f) || self.checkIfHallOrder(f, -1) || self.checkIfHallOrder(f, 1) {
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
}*/


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
    return sum(self.Commands) * stopTime + (abs(turn_floor - self.curFloor()) + abs(final_floor - turn_floor)) * travelTime;
}




func (self Orders) orderIsBestForMe(orderFloor int, orderDirection int) bool {
    lowestCost := 300
    bestCart_addr := ""
    for addr, cart := range self.Carts {
        //clientCost := self.costForClient(cart, orderFloor, orderDirection)
        addedCost := cart.cost(orderFloor, orderDirection) - cart.cost(cart.curFloor(), cart.curDir());//self.addedCostForElevator(cart, orderFloor, orderDirection)
        cost := /*clientCost*3*/ + addedCost*1
        fmt.Println("IP:", addr, "Cost:", cost);
        if cost < lowestCost || cost == lowestCost && addr < bestCart_addr {
            lowestCost = cost
            bestCart_addr = addr
        }
    }
    return bestCart_addr == self.Addr
}
/*
func (self Orders) GetDirection() int {
    curDir := self.Carts[self.Addr].curDir()
    curFloor := self.Carts[self.Addr].curFloor()
    iterateDirection := curDir
    if iterateDirection == 0 {
        iterateDirection = 1
    }
    floor := curFloor + iterateDirection;
    for floor < N_FLOORS {
        if floor < 0 {
            break
        } else if self.Carts[self.Addr].checkIfCommand(floor) {
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
        } else if self.Carts[self.Addr].checkIfCommand(floor) {
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
}*/

func (self Orders) GetDirection() int {
    direction := self.CurDir(self.Addr);
    if direction == 0 {
        direction = 1
    }
    floor := self.CurFloor(self.Addr);
    if  self.search_for_orders_in_direction(floor, direction) || self.Carts[self.Addr].furthest_command(floor, direction) != floor {
        return direction;
    } else if self.search_for_orders_in_direction(floor, -direction) || self.Carts[self.Addr].furthest_command(floor, -direction) != floor {
        return -direction;
    } else {
        return 0;
    }
}

func (self *Orders) Sync(s *Orders, light_channel chan<- Order) {
    for f := range s.Hall {
        for b := range s.Hall[f] {
            if s.Hall[f][b] {
                light_channel <-Order{Button: b, Floor: f, Value: true};
            }
            s.Hall[f][b] = self.Hall[f][b] || s.Hall[f][b];
            self.Hall[f][b] = s.Hall[f][b]
        }
    }
    if self.Carts == nil {
        self.Carts = make(map[string]*cart)
    }
    for addr, cart := range s.Carts {
        if addr != self.Addr {
            self.Carts[addr] = cart;
        }
    }
    s.Carts[self.Addr] = self.Carts[self.Addr];
}
