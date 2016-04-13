package io
/*
 #cgo CFLAGS: -std=c99
 #cgo LDFLAGS: -lcomedi -lm
 #include "io.h"
*/
import "C"

func Init() int {
	return int(C.io_init());
}
func SetBit(channel int) {
	C.io_set_bit(C.int(channel));
}
func ClearBit(channel int) {
	C.io_clear_bit(C.int(channel));
}

func ReadBit(channel int) bool {
	return int(C.io_read_bit(C.int(channel))) != 0;
}
func ReadAnalog(channel int) int {
	return int(C.io_read_analog(C.int(channel)));
}
func WriteAnalog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value));
}