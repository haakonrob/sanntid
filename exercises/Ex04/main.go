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
	runtime.GOMAXPROCS(runtime.NumCPU())
	portnr := ":20023"
	passcode := "svekonrules"
	
	go UDPserver(portnr, passcode)
	for{
		broadcastLocalAddressUDP("129.241.187.255"+portnr, passcode)
		time.Sleep(time.Second*2)
	}
}


func UDPserver(port string, passcode string){
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
			fmt.Println(msg)		
		} else {
			reader.Reset(sockln)		
		}
	}
}


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

