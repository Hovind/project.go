package obj

import (
    "net"
)

const (
    N_FLOORS     = 4
    N_BUTTONS    = 3
    N_DIRECTIONS = 2
)

type Order struct {
    Button, Floor int
    Value bool
}

type Message struct {
    Code int
    Body []byte
    Origin, Target *net.UDPAddr
}

func NewMessage(code int, body []byte, local_addr, target_addr *net.UDPAddr) *Message {
    return &Message{Code: code, Body: body, Origin: local_addr, Target: target_addr};
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
