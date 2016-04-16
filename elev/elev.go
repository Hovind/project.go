package elev
import (
    "project.go/io"
    . "project.go/obj"
)



const (
    DOWN                = -1 + iota
    STOP
    UP
)

const (
    N_FLOORS            = 4
    N_BUTTONS           = 3
    MOTOR_SPEED         = 2800

    MOTOR               = (0x100+0)

    BUTTON_DOWN2        = (0x200+0)
    BUTTON_UP3          = (0x200+1)
    BUTTON_DOWN3        = (0x200+2)
    BUTTON_DOWN4        = (0x200+3)
    SENSOR_FLOOR1       = (0x200+4)
    SENSOR_FLOOR2       = (0x200+5)
    SENSOR_FLOOR3       = (0x200+6)
    SENSOR_FLOOR4       = (0x200+7)

    LIGHT_FLOOR_IND1    = (0x300+0)
    LIGHT_FLOOR_IND2    = (0x300+1)
    LIGHT_DOOR_OPEN     = (0x300+3)
    LIGHT_DOWN4         = (0x300+4)
    LIGHT_DOWN3         = (0x300+5)
    LIGHT_UP3           = (0x300+6)
    LIGHT_DOWN2         = (0x300+7)
    LIGHT_UP2           = (0x300+8)
    LIGHT_UP1           = (0x300+9)

    LIGHT_COMMAND4      = (0x300+10)
    LIGHT_COMMAND3      = (0x300+11)
    LIGHT_COMMAND2      = (0x300+12)
    LIGHT_COMMAND1      = (0x300+13)
    LIGHT_STOP          = (0x300+14)
    MOTOR_DIR           = (0x300+15)
    BUTTON_UP2          = (0x300+16)
    BUTTON_UP1          = (0x300+17)
    BUTTON_COMMAND4     = (0x300+18)
    BUTTON_COMMAND3     = (0x300+19)

    BUTTON_COMMAND2     = (0x300+20)
    BUTTON_COMMAND1     = (0x300+21)
    BUTTON_STOP         = (0x300+22)
    OBSTRUCTION         = (0x300+23)

    // Matrix symmetry
    BUTTON_DOWN1        = -1
    BUTTON_UP4          = -1
    LIGHT_DOWN1         = -1
    LIGHT_UP4           = -1
)

var LAMP_CHANNEL_MATRIX = [][] int {
    {LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
    {LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
    {LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
    {LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4}};

var BUTTON_CHANNEL_MATRIX = [][] int {
    {BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
    {BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4}};

func Init() {
    io.Init();

    for f := 0; f < N_FLOORS; f++ {
        for b := 0; b < N_BUTTONS; b++ {
            SetButtonLamp(b, f, false);
        }
    }
    SetStopLamp(false);
    SetDoorOpenLamp(false);
    SetFloorIndicator(0);
}

func CheckHardware(localOrderCh chan Order, floorSignalCh chan int, stopSignalCh chan bool) {
    for {
        checkOrderButtons(localOrderCh);
        checkFloorSensors(floorSignalCh);
        checkStopSignal(stopSignalCh);
    }
}

func SetMotorDirection(dirn int) {
    if dirn == 0 {
        io.WriteAnalog(MOTOR, 0);
    } else if dirn > 0 {
        io.ClearBit(MOTOR_DIR);
        io.WriteAnalog(MOTOR, MOTOR_SPEED);
    } else if dirn < 0 {
        io.SetBit(MOTOR_DIR);
        io.WriteAnalog(MOTOR, MOTOR_SPEED);
    }
}


func SetButtonLamp(button int, floor int, value bool) {
    if (value) {
        io.SetBit(LAMP_CHANNEL_MATRIX[floor][button]);
    } else {
        io.ClearBit(LAMP_CHANNEL_MATRIX[floor][button]);
    }
}


func SetFloorIndicator(floor int) {
    // Binary encoding. One light must always be on.
    if (floor & 0x02 != 0) {
        io.SetBit(LIGHT_FLOOR_IND1);
    } else {
        io.ClearBit(LIGHT_FLOOR_IND1);
    }

    if (floor & 0x01 != 0) {
        io.SetBit(LIGHT_FLOOR_IND2);
    } else {
        io.ClearBit(LIGHT_FLOOR_IND2);
    }
}

func SetDoorOpenLamp(value bool) {
    if (value) {
        io.SetBit(LIGHT_DOOR_OPEN);
    } else {
        io.ClearBit(LIGHT_DOOR_OPEN);
    }
}


func SetStopLamp(value bool) {
    if (value) {
        io.SetBit(LIGHT_STOP);
    } else {
        io.ClearBit(LIGHT_STOP);
    }
}



func getButtonSignal(button int, floor int) bool {
    return io.ReadBit(BUTTON_CHANNEL_MATRIX[floor][button]);
}

func getFloorSensorSignal() int {
    if (io.ReadBit(SENSOR_FLOOR1)) {
        return 0;
    } else if (io.ReadBit(SENSOR_FLOOR2)) {
        return 1;
    } else if (io.ReadBit(SENSOR_FLOOR3)) {
        return 2;
    } else if (io.ReadBit(SENSOR_FLOOR4)) {
        return 3;
    } else {
        return -1;
    }
}

func checkOrderButtons(localOrderCh chan Order) {
    for floor := 0; floor < N_FLOORS; floor++ {
        for button := 0; button < N_BUTTONS; button++ {
            if getButtonSignal(button, floor) {
                localOrderCh <- Order{Button: button, Floor: floor};
            }
        }
    }
}

func Button_checker() <-chan Order {
    local_order_channel := make(chan Order);
    go func() {
        previous_button_signal := [4][3]bool{};
        for {
            for floor := 0; floor < N_FLOORS; floor++ {
                for button := 0; button < N_BUTTONS; button++ {
                    signal := getButtonSignal(button, floor);
                    if signal && !previous_button_signal[floor][button] {
                        local_order_channel <-Order{Button: button, Floor: floor, Value: true};
                    }
                    previous_button_signal[floor][button] = signal;
                }
            }
        }
    }();
    return local_order_channel;
}

func Floor_checker() <-chan int {
    floor_sensor_channel := make(chan int);
    go func() {
        prev_floor := -1;
        for {
            floor := getFloorSensorSignal();
            if floor != -1 && floor != prev_floor {
                floor_sensor_channel <-floor;
            }
            prev_floor = floor;
        }
    }();
    return floor_sensor_channel;
}


func Stop_checker() <-chan bool {
    stop_signal_channel := make(chan bool);
    go func() {
        prev_stop_signal := false;
        for {
            stop_signal := io.ReadBit(BUTTON_STOP);
            if stop_signal && !prev_stop_signal {
                stop_signal_channel <- stop_signal;
            }
            prev_stop_signal = stop_signal;
        }
    }();
    return stop_signal_channel;

}

func Light_manager() chan<- Order {
    light_channel := make(chan Order);
    go func() {
        for {
            order := <-light_channel;
            SetButtonLamp(order.Button, order.Floor, order.Value);
        }
    }();
    return light_channel;
}
