package driver

/*
#include<io.h>
*/

import "C"


func ioInit() int{
	return C.io_init()
}
//int io_init(void);


func ioSetBit(ch int){
	C.io_set_bit(ch)
}
//void io_set_bit(int channel);

func ioClearBit(ch int){
	C.io_clear_bit(ch)
}
//void io_clear_bit(int channel);


func ioReadBit(ch int)int{
	return C.io_read_bit(ch)
}
//int io_read_bit(int channel);

func ioReadAnalog(ch int)int{
	return C.io_read_analog(ch)
}
//int io_read_analog(int channel);


func ioWriteAnalog(ch int, value int){
	C.io_read_analog(ch, value)
}
//void io_write_analog(int channel, int value);
