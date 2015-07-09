// Copyright 2015 go-sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"net"
)

type Packet struct {
	addr *net.UDPAddr
	data []byte
}

const UDP_PACKET_SIZE = 2048

func Start(address string) *net.UDPConn {
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
