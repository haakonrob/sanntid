package localnet

import (
	"net"
	"strings"
	"sort"
	"time"
	"fmt"
)

var (
	localIP string
	broadcastIP string
	IPList [] string
	IPTimestamps map[string]time.Time
)

func Init(){
	IP()
	BroadcastIP()
	IPList = make([]string,0,20)
	IPTimestamps = make(map[string]time.Time)

}

func IP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func BroadcastIP() (string, error) {
	if broadcastIP == "" {
		if localIP == "" {
			IP()
		}
		temp := strings.Split(localIP, ".")
		broadcastIP = temp[0]+"."+temp[1]+"."+temp[2]+".255"
	}
	return broadcastIP, nil
}
/*
func KnownIPs()([]string, error){
	return IPList, nil
}
/*
func GetNumberOfNodes()(int){
	return len(IPList) +1
}
*/
func PeerUpdate(IPPing string){
	for i:=0; i<len(IPList); i++ {
		if IPList[i] == IPPing {
			IPTimestamps[IPPing] = time.Now()
			return
		} 
	}
	// If IP is not in list:
	IPList = append(IPList, IPPing)
	sort.Strings(IPList)
	IPTimestamps[IPPing] = time.Now()
}
/*
func NewNode(newIP string)(error){
	for i:=0; i<len(IPList); i++ {
		if IPList[i] == newIP {
			return errors.New("IP is already in list")
		} 
	}
	IPList = append(IPList, newIP)
	sort.Strings(IPList)
	return nil
}
/*
func RemoveNode(IP string)(error){
	for i:=0; i<len(IPList); i++ {
		if IPList[i] == IP {
			IPList = append(IPList[:i], IPList[i+1:]...)
			return nil
		} 
	}
	return errors.New("IP to delete is not in list")
}
*/
func RemoveDeadConns(timeout time.Duration)(bool){
	updateNeeded := false
	for i:=0; i<len(IPList)-1; i++ {
		timestamp, IPexists := IPTimestamps[IPList[i]]
		if IPexists {
			if time.Since(timestamp) > timeout {
				delete(IPTimestamps, IPList[i])
				IPList = append(IPList[:i], IPList[i+1:]...)
				updateNeeded = true
			}
		} 
	}
	return updateNeeded
}

func NextNode()(string){
	if len(IPList) == 0 {
		fmt.Println("No IPs")
		return "bad"
	} 
	if len(IPList) == 1 {
		return IPList[0]
	} 
	// smallest member
	if localIP < IPList[0] {
		return IPList[0]
	}
	// somewhere inbetween
	for i:=0;i<len(IPList)-1;i++ {
		if localIP == IPList[i] {
			return "BadIPlist" //shouldn't happen
		} else if localIP > IPList[i] && localIP < IPList[i+1] {
			return IPList[i+1]
		} 
	}
	// reached end of list, wrap around
	return IPList[0]
}
func NodeNumber()(int){
	
	for i:=0;i<len(IPList):i++{
		if localIP < IPList[i] {
			return i
		}
	} 
	return i
	
}

