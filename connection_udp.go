// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"log"
	"net"
)

// Packet stores UDP connection data and payload
type Packet struct {
	addr *net.UDPAddr
	data []byte
}

// MaxPayload sets the largest accepted UDP packet
const udpPacketSize = 2048

// StartUDP prepares and returns UDP connection
func StartUDP(address string) (chan Packet, chan Packet) {
	addr, err := net.ResolveUDPAddr("udp", address)
	CheckConnError(err)

	conn, err := net.ListenUDP("udp", addr)
	CheckConnError(err)

	// Outbound channel uses connection to send messages
	outbound := make(chan Packet)
	// Inbound channel passes received message to handleIncomingPacket function
	inbound := make(chan Packet)

	// Goroutine for receiving messages and passing them to handleIncomingPacket function
	go recvUDP(conn, inbound)
	// Goroutine for sending messages
	go sendUDP(conn, outbound)

	return inbound, outbound
}

func sendUDP(connection *net.UDPConn, outbound chan Packet) {
	for packet := range outbound {
		_, err := connection.WriteToUDP(packet.data, packet.addr)
		if err != nil {
			log.Println("Error on write: ", err)
			continue
		}
	}
}

func recvUDP(connection *net.UDPConn, inbound chan Packet) {
	for {
		b := make([]byte, udpPacketSize)
		n, addr, err := connection.ReadFromUDP(b)
		if err != nil {
			log.Println("Error on read: ", err)
			continue
		}
		inbound <- Packet{addr, b[:n]}
	}
}
