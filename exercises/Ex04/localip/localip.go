package localip

import (
	"net"
	"strings"
)

var localIP string
var broadcastIP string

func Get() (string, error) {
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

func GetBroadcast() (string, error) {
	if broadcastIP == "" {
		if localIP == "" {
			Get()
		}
		temp := strings.Split(localIP, ".")
		broadcastIP = temp[0]+"."+temp[1]+"."+temp[2]+".255"
	}
	return broadcastIP, nil
}
