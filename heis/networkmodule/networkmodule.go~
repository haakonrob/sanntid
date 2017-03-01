package networkmodule

import(
	_"fmt"
	"./ringnode"
	"./networkmonitor"
)

const (
	TCPPortIn = ":20024"
	TCPPortOut = ":20025"
	UDPPort = ":20023"
	UDPPasscode = "svekonrules"
)

func Init(packetChannel chan string, monitorChannel chan string){
	updateChannel := make(chan string)
	
	go ringnode.RingNode(packetChannel, updateChannel, TCPPortIn, TCPPortOut)
	networkmonitor.JoinNetwork(monitorChannel, updateChannel, UDPPort, UDPPasscode)
	return
}

