package main

import (
	"fmt"
	"net"
	"time"
	"runtime"
	"errors"
	"os"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	go TCPserver(":20023")
	//go UDPserver(":20012")
	for {
		//UDPclient("129.241.187.43:20012")
		TCPclient("129.241.187.54:20023")
		time.Sleep(time.Second*3)
				
	}
}

func TCPclient(targetAddr string){
	/*
	addrLocal, err := net.ResolveTCPAddr("tcp", "129.241.187.255"+":34933")
	CheckError(err, " ")
	*/

	addrServer, err := net.ResolveTCPAddr("tcp",targetAddr)
	CheckError(err, " ")		

	conn, err := net.DialTCP("tcp", nil, addrServer)
	CheckError(err, " ")
	
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	fmt.Println("Received: ", string(buf[0:n]), "\n")	
	

	b := []byte("heihei" + "" +"")
	_, err = conn.Write(b)	
	CheckError(err, "TCPwrite")

	n, err = conn.Read(buf)
	fmt.Println("Received: ", string(buf[0:n]), "\n")

	conn.Close()
}

func TCPserver(port string){
	addr, err := net.ResolveTCPAddr("tcp", port)
	CheckError(err, " ")
	
	ln, err := net.ListenTCP("tcp", addr)
	CheckError(err, " ")

	sock, err := ln.Accept()
	CheckError(err, " ")
	
	msg := []byte("Welcome to the Pleasuredrome..."+"")
	sock.Write(msg)
	
	buf := make([]byte, 1024)
	for {
		n, err := sock.Read(buf)
		CheckError(err, " ")
		fmt.Println("Received: ", string(buf[0:n]), "\n")		
	}

	
	
}

func UDPserver(port string){
	addr, err := net.ResolveUDPAddr("udp", port)
    CheckError(err, " ")

	sockln, err := net.ListenUDP("udp", addr)
	CheckError(err, " ")

	buf := make([]byte, 1024)

	for {
		n, addr, err := sockln.ReadFromUDP(buf)
		CheckError(err, " ")
		fmt.Println("Received: ", string(buf[0:n]), "\nAddress: ", addr, "\n")		
	}
}

func UDPclient(targetAddr string){
	
	addr, err := getLocalIP()
	CheckError(err, " ")

	_ = addr	

	addrLocal, err := net.ResolveUDPAddr("udp", "129.241.187.54"+":20012")
	CheckError(err, " ")
	_ = addrLocal
	addrServer, err := net.ResolveUDPAddr("udp",targetAddr)
	CheckError(err, " ")
	

	conn, err := net.DialUDP("udp", nil, addrServer)
	CheckError(err, "Dial")
	
	b := []byte("heihei" + "")
	_, err = conn.Write(b)
	fmt.Println("b: ", string(b))
	CheckError(err, "write")
	conn.Close()
	/*
	//conn.Close()
	
	//sockln, err := net.ListenUDP("udp", addrLocal)
	//CheckError(err, " ")

	n, addr2, err := conn.ReadFromUDP(buf)
	fmt.Println("Received: ", string(buf[0:n]), "\nAddress: ", addr2, "\n")
	conn.Close()*/
}

func getLocalIP() (string, error) {
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





