package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include<io.h>
*/

import "C"


func ioInit() bool{
	return int(C.io_init()) == 1
}
//int io_init(void);


func ioSetBit(ch int){
	C.io_set_bit(C.int(ch))
}
//void io_set_bit(int channel);

func ioClearBit(ch int){
	C.io_clear_bit(C.int(ch))
}
//void io_clear_bit(int channel);


func ioReadBit(ch int)int{
	return int(C.io_read_bit(C.int(ch)))
}
//int io_read_bit(int channel);

func ioReadAnalog(ch int)int{
	return int(C.io_read_analog(C.int(ch)))
}
//int io_read_analog(int channel);


func ioWriteAnalog(ch int, value int){
	C.io_read_analog(C.int(ch), C.int(value))
}
//void io_write_analog(int channel, int value);