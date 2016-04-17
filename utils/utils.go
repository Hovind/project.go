package utils

import (
	"time"
	. "project.go/obj"
)


type Timer struct {
    Timer *time.Timer
    Running bool
}

func NewTimer() *Timer {
	t := time.NewTimer(0);
	<-t.C;
	return &Timer{Timer: t, Running: false};
}

func (t *Timer) Start(d time.Duration) {
	t.Timer.Reset(d);
	t.Running = true;
}

func (t *Timer) Stop() {
	t.Timer.Stop();
	t.Running = false;
}


func Abs(x int) int {
    if x < 0 {
        return -x
    } else {
        return x
    }
}

func Sum(commands [N_FLOORS]bool) int {
    sum := 0;
    for _, e := range commands {
        if e {
            sum += 1;
        }
    }
    return sum;
}
func Max(a, b int) int {
    if a > b {
        return a;
    } else {
        return b;
    }
}

func Sign(a int) int {
    if a > 0 {
        return 1;
    } else if a < 0 {
        return -1;
    } else {
        return 0;
    }
}
