package main

import (
	"fmt"
	"net"
	"time"
	"runtime"
	"errors"
	"os"
	"bufio"
)

func main(){
	//runtime.GOMAXPROCS(runtime.NumCPU())

	connectedIPs := make([]string, 0, 20)
	_ = connectedIPs
	localIP, err := getLocalIP()	
	CheckError(err, "No local IP")

	localBroadcast := "129.241.187.255"
	portnr := ":20022"
	passcode := "svekonrules"
	msg := passcode+"\n"+"hello"+"\n"
	
	chanUDP := make(chan string, 1)

	go listenUDP(chanUDP, localIP, portnr, passcode)

	for{
		/*
			if connection false, sendUDP(), sleep
			read listenUDP() channel, add IP addresses to list
			while list != empty, sort IP address into ring and connect
				sort IP address list
				lower neighbour disconnects upper socket
		*/

		sendUDP(localBroadcast+portnr, msg)
		imsg := <-chanUDP
		//unconnectedIPs = append(unconnectedIPs, imsg)
		//fmt.Println(unconnectedIPs)
		if len(connectedIPs) == 0 {
			connectedIPs = append(connectedIPs, imsg)
		} else {
			flag := true
			for i:=0; i<len(connectedIPs); i++ {
				if connectedIPs[i] == imsg {
					flag = false
					break
				} 
			}
			if flag {
				// connect to IP
				connectedIPs = append(connectedIPs, imsg)
				fmt.Println(connectedIPs)	
			}
		}
		fmt.Println(connectedIPs)
		time.Sleep(time.Second*2)
		
		
	}
}


func listenUDP(chanUDP chan string, localIP string, port string, passcode string){
	addr, err := net.ResolveUDPAddr("udp", port)
    	CheckError(err, " ")

	sockln, err := net.ListenUDP("udp", addr)
	CheckError(err, " ")

	//buf := make([]byte, 1024)
	reader := bufio.NewReader(sockln)

	for {
		code, err := reader.ReadString('\n')
		CheckError(err, "reader.ReadString")
		//fmt.Println("read:" + code)

		if code == (passcode + "\n") {
			msg, err := reader.ReadString('\n')
			CheckError(err, "reader.ReadString")
			// ignore computer's own messages
			if msg != (localIP + "\n") {	
				chanUDP <- msg[:len(msg)-1]
			}		
		} else {
			reader.Reset(sockln)		
		}
	}
}

func sendUDP(target string, msg string){
	
	addr, err := net.ResolveUDPAddr("udp",target)
	CheckError(err, "Resolve server addr")
	
	conn, err := net.DialUDP("udp", nil, addr)
	CheckError(err, "UDPDial")
	
	b := []byte(msg)

	_, err = conn.Write(b)
	//fmt.Println("Sent: ", string(b))
	CheckError(err, "conn.Write()")
	conn.Close()
}
/*
func broadcastLocalAddressUDP(targetAddr string, passcode string){
	
	localIP, err := getLocalIP()
	CheckError(err, " ")

	addrServer, err := net.ResolveUDPAddr("udp",targetAddr)
	CheckError(err, "Resolve server addr")
	
	conn, err := net.DialUDP("udp", nil, addrServer)
	CheckError(err, "UDPDial")
	
	b := []byte(passcode + "\n" + localIP + "\n")

	_, err = conn.Write(b)
	//fmt.Println("Sent: ", string(b))
	CheckError(err, "conn.Write()")
	conn.Close()
}
*/
func getLocalIP() (string, error) {
// Find source of this
ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}


func CheckError(err error, place string){
	if err != nil {
		fmt.Println("Error detected at: ", place)
		fmt.Println(err)
		os.Exit(0)
	}
}

