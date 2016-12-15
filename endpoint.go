// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"net"
)

// Packet stores UDP connection data and payload
type Packet struct {
	addr *net.UDPAddr
	data []byte
}

// MaxPayload sets the largest accepted UDP packet
const MaxPayload = 2048

// Start prepares and returns UDP connection
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
