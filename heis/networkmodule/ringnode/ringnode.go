package ringnode

import (
	"net"
)

var (
	nextNode net.Conn
	prevNode net.Conn
)

func Init(elevChannel, updateChannel){}
