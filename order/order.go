package order

import (
    . "project.go/obj"
    "project.go/utils"
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
    Commands         [N_FLOORS]bool
    Active bool
}

type Orders struct {
    Addr string
    Hall       [N_FLOORS][N_DIRECTIONS]bool
    Carts      map[string]*cart
}


//SAVED FOR LATER
func NewOrders(local_addr string) *Orders {
    o := &Orders{Addr: local_addr, Carts: make(map[string]*cart)}
    o.Carts[local_addr] = &cart{Floor: 0, Direction: 0}
    return o
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


func (self Orders) CheckFloorAction(orderFloor int, orderDirection int) int {
    if  self.Carts[self.Addr].checkIfCommand(orderFloor) ||
        self.checkIfHallOrder(orderFloor, orderDirection) ||
        self.checkIfHallOrder(orderFloor, -orderDirection) &&
        !self.search_for_orders_in_direction(orderFloor, orderDirection) &&
        self.Carts[self.Addr].furthest_command(orderFloor, orderDirection) == orderFloor {
        return OPEN_DOOR;
    } else if !self.search_for_orders_in_direction(orderFloor, orderDirection) && self.Carts[self.Addr].furthest_command(orderFloor, orderDirection) == orderFloor {
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


func (self cart) cost(floor, direction int) int {
    turn_floor := self.curFloor();
    final_floor := floor;
    if self.curDir() == utils.Sign(floor - self.curFloor()) {
        turn_floor = self.curFloor() + self.curDir() * utils.Max(utils.Abs(self.furthest_command(self.curFloor(), self.curDir()) - self.curFloor()), utils.Abs(floor - self.curFloor()))
        final_floor = self.furthest_command(turn_floor, -self.curDir());
    } else if self.curDir() == -utils.Sign(floor - self.curFloor()) {
        turn_floor = self.furthest_command(self.curFloor(), self.curDir());
        final_floor = turn_floor - self.curDir() * utils.Max(utils.Abs(self.furthest_command(turn_floor, -self.curDir()) - turn_floor), utils.Abs(floor - turn_floor));
    }
    return utils.Sum(self.Commands) + (utils.Abs(turn_floor - self.curFloor()) + utils.Abs(final_floor - turn_floor));
}




func (self Orders) orderIsBestForMe(orderFloor int, orderDirection int) bool {
    min_weighted_cost := 0
    best_cart_addr := ""
    for addr, cart := range self.Carts {
        previous_cost := cart.cost(cart.curFloor(), cart.curDir());
        cost := cart.cost(orderFloor, orderDirection)
        cost_difference :=  cost - previous_cost;
        weighted_cost := cost_difference;
        if best_cart_addr == "" || weighted_cost < min_weighted_cost || weighted_cost == min_weighted_cost && addr < best_cart_addr {
            min_weighted_cost = cost
            best_cart_addr = addr
        }
    }
    return best_cart_addr == self.Addr
}

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

func (self *Orders) Sync(s *Orders, new_order *bool, light_channel chan<- Order) {
    for f := range s.Hall {
        for b := range s.Hall[f] {
            if s.Hall[f][b] {
                light_channel <-Order{Button: b, Floor: f, Value: true};
                *new_order = true;
            }
            s.Hall[f][b] = self.Hall[f][b] || s.Hall[f][b];
            self.Hall[f][b] = s.Hall[f][b]
        }
    }
    c := self.Carts[self.Addr]
    self.Carts = make(map[string]*cart)
    for addr, cart := range s.Carts {
        self.Carts[addr] = cart;
    }
    self.Carts[self.Addr] = c;
    s.Carts[self.Addr] = c;
}
