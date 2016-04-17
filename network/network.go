package network

import (
    "fmt"
    "net"
    "strings"
    "encoding/json"
    "time"
    "project.go/utils"
    "project.go/order"
    . "project.go/obj"
)

func addr_is_remote(local_addrs []net.Addr, addr *net.UDPAddr) bool {
    for _, local_addr := range local_addrs {
        if strings.Contains(local_addr.String(), addr.IP.String()) {
            return false;
        }
    }
    return true;
}

func resolve_local_addr(broadcast_addr *net.UDPAddr, broadcast_port string) *net.UDPAddr {
    temp_socket, _ := net.DialUDP("udp4", nil, broadcast_addr);
    defer temp_socket.Close();
    temp_addr, _ := net.ResolveUDPAddr("udp4", temp_socket.LocalAddr().String());
    local_addr, _ := net.ResolveUDPAddr("udp4", temp_addr.IP.String() + ":" + broadcast_port);
    return local_addr;
}

func listening_worker(socket *net.UDPConn, local_addr *net.UDPAddr) (chan Message, <-chan Message) {
    local_addrs, _ := net.InterfaceAddrs();

    from_network_channel := make(chan Message, 10);
    rcv_channel := make(chan Message);
    go func() {
        b := make([]byte, 1024);
        for {
            n, addr, err := socket.ReadFromUDP(b);
            if err != nil {
                fmt.Println("Could not read from UDP:", err.Error());
            } else {
                msg := Message{};
                err := json.Unmarshal(b[:n], &msg);
                if err != nil {
                    fmt.Println("Could not unmarshal message.");
                } else if addr_is_remote(local_addrs, addr) && n > 0 {
                    if msg.Code <= SYNC {
                        from_network_channel<-msg;
                    }
                    if msg.Origin.IP.String() != local_addr.IP.String() {
                        fmt.Println("Received message with code", msg.Code, "with body", msg.Body, "from", addr.String());
                        rcv_channel <-msg;
                    } else {
                        rcv_channel <-*NewMessage(KEEP_ALIVE, nil, nil, nil);
                    }
                }
            }
        }
    }();
    return from_network_channel, rcv_channel;
}

func send(msg Message, socket *net.UDPConn, addr *net.UDPAddr) (error) {
    b, err := json.Marshal(msg);
    if err != nil {
        fmt.Println("Could not marshal message.");
        return err;
    }
    _, err = socket.WriteToUDP(b, addr);
    if err != nil {
        fmt.Println("Could not send:", err.Error());
    } else {
        fmt.Println("Sent message with code", msg.Code, "and body", msg.Body, "to:", addr.String());
    }
    return err;
}

func request_head(local_addr, broadcast_addr *net.UDPAddr, socket *net.UDPConn) error {
    msg := *NewMessage(HEAD_REQUEST, nil, local_addr, nil);
    return send(msg, socket, broadcast_addr);
}

func find_network(local_addr, broadcast_addr *net.UDPAddr, socket *net.UDPConn, to_network_channel, from_network_channel chan Message, rcv_channel <-chan Message) *net.UDPAddr {
    desync(local_addr, from_network_channel);
    request_head(local_addr, broadcast_addr, socket);
    for {
        select {
        case msg := <-rcv_channel:
            switch msg.Code {
            case TAIL_REQUEST, HEAD_REQUEST, CONNECTION:
                head_addr := msg.Origin;
                if msg.Code != CONNECTION {
                    msg = *NewMessage(CONNECTION, nil, local_addr, head_addr);
                }
                send(msg, socket, head_addr);
                return head_addr;
        }
        case msg := <-to_network_channel:
            msg.Origin = local_addr;
            from_network_channel <-msg;
        case <-time.After(5 * time.Second):
            request_head(local_addr, broadcast_addr, socket);
        }
    }
}

func desync(local_addr *net.UDPAddr, from_network_channel chan<- Message) error {
    orders := *order.NewOrders(local_addr.IP.String());
    b, err := json.Marshal(orders);
    if err != nil {
        return err;
    }
    from_network_channel <-*NewMessage(SYNC, b, local_addr, nil);
    return nil;
}

func send_sync(local_addr, head_addr *net.UDPAddr, socket *net.UDPConn) error {
    orders := *order.NewOrders(local_addr.IP.String());
    b, err := json.Marshal(orders);
    if err != nil {
        return err;
    }
    msg := *NewMessage(SYNC, b, local_addr, nil);
    return send(msg, socket, head_addr);
}

func maintain_network(local_addr, head_addr *net.UDPAddr, socket *net.UDPConn, to_network_channel, rcv_channel <-chan Message) {
    tail := utils.NewTimer();
    keep_alive := utils.NewTimer();
    send_sync(local_addr, head_addr, socket);
    for {
        select {
        case msg := <-to_network_channel:
            msg.Origin = local_addr;
            send(msg, socket, head_addr);
        case msg :=  <-rcv_channel:
            tail.Stop();
            keep_alive.Start(2 * time.Second);
            switch msg.Code {
            case KEEP_ALIVE, TAIL_REQUEST, SYNC:
                break;
            case CONNECTION:
                if head_addr == nil || head_addr.String() == msg.Target.String() {
                   head_addr = msg.Origin;
                }
                send(msg, socket, head_addr);
            case HEAD_REQUEST:
                addr := msg.Origin;
                msg := *NewMessage(TAIL_REQUEST, []byte{}, local_addr, nil);
                send(msg, socket, addr);
            case TAIL_DEAD:
                fmt.Println("Cycle broken.");
                sleep_multiplier := int(msg.Body[0]) + 1;
                msg.Body[0] = byte(sleep_multiplier);
                send(msg, socket, head_addr);
                head_addr = nil;
                fmt.Println("Sleeping for", sleep_multiplier, "seconds.");
                time.Sleep(time.Duration(sleep_multiplier)*time.Second);
                fmt.Println("Done sleeping.");
                return;
            default:
                send(msg, socket, head_addr);
            }
        case <-tail.Timer.C:
            fmt.Println("Breaking cycle.");
            sleep_multiplier := 1;
            msg := *NewMessage(TAIL_DEAD, []byte{byte(sleep_multiplier)}, local_addr, nil);
            send(msg, socket, head_addr);
            head_addr = nil;
            fmt.Println("Sleeping for", sleep_multiplier, "seconds.");
            time.Sleep(time.Duration(sleep_multiplier)*time.Second);
            fmt.Println("Done sleeping.");
            return;
        case <-keep_alive.Timer.C:
            msg := *NewMessage(KEEP_ALIVE, nil, local_addr, nil);
            send(msg, socket, head_addr);
            if !tail.Running {
                fmt.Println("Starting timer.")
                tail.Start(10 * time.Second);
            }
        }
    }
}

func Manager(broadcast_port string) (string, chan<- Message, <-chan Message) {
    broadcast_addr, _ := net.ResolveUDPAddr("udp4", net.IPv4bcast.String() + ":" + broadcast_port);
    local_addr := resolve_local_addr(broadcast_addr, broadcast_port);

    listen_addr, _ := net.ResolveUDPAddr("udp4", ":" + broadcast_port);
    socket, err := net.ListenUDP("udp4", listen_addr);
    if err != nil {
        fmt.Println("Could not create socket:", err.Error());
        return "", nil, nil;
    } else {
        fmt.Println("Socket has been created:", socket.LocalAddr().String());
    }

    from_network_channel, rcv_channel := listening_worker(socket, local_addr);
    to_network_channel := make(chan Message, 10);
    go func() {
        for {
            head_addr := find_network(local_addr, broadcast_addr, socket, to_network_channel, from_network_channel, rcv_channel);
            maintain_network(local_addr, head_addr, socket, to_network_channel, rcv_channel);
        }
    }();
    return local_addr.IP.String(), to_network_channel, from_network_channel;
}
