package order

import (
	"fmt"
	//"math"
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

type cart struct {
	floor, direction int
	commands         [4]bool
}

type Orders struct {
	local_addr string
	hall       [4][2]bool
	carts      map[string]*cart
}

//SAVED FOR LATER
func NewOrders(local_addr string) *Orders {
	o := &Orders{local_addr: local_addr, carts: make(map[string]*cart)}
	o.carts[local_addr] = &cart{floor: -1, direction: -1}
	return o
}

func (self Orders) Print() {
	fmt.Println("\nLocal address:", self.local_addr)
	fmt.Println("Hall", self.hall)
	for key, value := range self.carts {
		fmt.Println("IP:", key)
		fmt.Println("Cart:", value.commands)
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

func (self Orders) CurFloor() int {
	return self.carts[self.local_addr].curFloor()
}

func (self *Orders) SetDir(addr string, direction int) {
	self.carts[self.local_addr].setDir(direction)
}

func (self *Orders) SetFloor(addr string, floor int) {
	self.carts[self.local_addr].setFloor(floor)
}



func (self *Orders) SetHallOrder(floor int, button int, value bool) {
	self.hall[floor][button] = value
}

func (self Orders) checkIfHallOrder(floor, direction int) bool {
	if direction == -1 {
		direction = 1
	} else {
		direction = 0
	}
	return self.hall[floor][direction]
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

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
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
		} else if floor == orderFloor && orderDirection == cart.curDir() {
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
				return cart.curFloor() - orderFloor
			}
			return (abs(cart.curFloor()-turnFloor)+abs(floor-orderFloor)+N_FLOORS-1)*travelTime + stops*stopTime
		} else if floor == orderFloor && orderDirection == -cart.curDir() {
			return (abs(cart.curFloor()-turnFloor)+abs(floor-orderFloor))*travelTime + stops*stopTime
		} else if cart.checkIfCommand(floor) {
			stops += 1
			noStops = false
		}
		floor -= cart.curDir()
	}
	return 0
}

func (self Orders) orderIsBestForMe(orderFloor int, orderDirection int) bool {
	lowestCost := 300
	bestCart_addr := ""
	for addr, cart := range self.carts {
		clientCost := self.costForClient(cart, orderFloor, orderDirection)
		addedCost := self.addedCostForElevator(cart, orderFloor, orderDirection)
		cost := clientCost*1 + addedCost*1
		if cost < lowestCost {
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
	floor := curFloor + curDir
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
	floor = curFloor - curDir
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
