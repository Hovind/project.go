package obj

import (
    "net"
)

const (
    N_FLOORS     = 4
    N_BUTTONS    = 3
    N_DIRECTIONS = 2
)

var Button = struct {
    Up, Down, Command int
}{0, 1, 2}

var Direction = struct {
    Up, Down, Stop int
}{1, -1, 0}

type Order struct {
    Button, Floor int
    Value bool
}

type Message struct {
    Code int
    Body []byte
    Origin, Target *net.UDPAddr
}

type Cart struct {
    Floor, Direction int
    Commands         [N_FLOORS]bool
}

type Orders struct {
    Addr       string
    Hall       [N_FLOORS][N_DIRECTIONS]bool
    Carts      map[string]*Cart
}

func New_message(code int, body []byte, local_addr, target_addr *net.UDPAddr) *Message {
    return &Message{Code: code, Body: body, Origin: local_addr, Target: target_addr};
}

func New_cart() *Cart {
    return &Cart{}
}

func New_orders(local_addr string, hall [N_FLOORS][N_DIRECTIONS]bool, cart_map map[string]*Cart) *Orders {
    o := &Orders{Addr: local_addr, Hall: hall, Carts: cart_map}
    o.Carts[local_addr] = New_cart();
    return o
}


const (
    ORDER = 100 + iota
    FLOOR_UPDATE
    DIRECTION_UPDATE
    SYNC
)

const (
    KEEP_ALIVE = 200 + iota
)

const (
    CONNECTION = 300 + iota
    HEAD_REQUEST
    TAIL_REQUEST
)

const (
    TAIL_DEAD = 400 + iota
    CYCLE_BREAK
)
