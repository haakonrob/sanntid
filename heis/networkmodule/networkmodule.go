package networkmodule

import(
	"fmt"
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
	
	//ringnode.Init(packetChannel, updateChannel, TCPPortIn, TCPPortOut)
	NetworkMonitor.JoinNetwork(monitorChannel, updateChannel, UDPPort, UDPPasscode)
	return true
}

