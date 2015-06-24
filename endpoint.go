package sip

import (
	"net"
)

func Start(address string) (*net.UDPConn) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	return connection
}

	
