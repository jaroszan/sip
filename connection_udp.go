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
func StartUDP(lAddr string, rAddr string) (chan SipMessage, chan SipMessage) {
	listenAddr, err := net.ResolveUDPAddr("udp", lAddr)
	CheckConnError(err)

	remotePeerAddr, err := net.ResolveUDPAddr("udp", rAddr)
	CheckConnError(err)

	conn, err := net.ListenUDP("udp", listenAddr)
	CheckConnError(err)

	// Outbound channel uses connection to send messages
	outbound := make(chan SipMessage)
	// Inbound channel passes received message to handleIncomingPacket function
	inbound := make(chan SipMessage)

	// Goroutine for receiving messages and passing them to handleIncomingPacket function
	go recvUDP(conn, inbound)
	// Goroutine for sending messages
	go sendUDP(conn, outbound, remotePeerAddr)

	return inbound, outbound
}

func sendUDP(connection *net.UDPConn, outbound chan SipMessage, remotePeer *net.UDPAddr) {
	for packet := range outbound {
		_, err := connection.WriteToUDP(SerializeSipMessage(packet), remotePeer)
		if err != nil {
			log.Println("Error on write: ", err)
			continue
		}
	}
}

//TODO check if datagrams are coming from predefined remote peer
func recvUDP(connection *net.UDPConn, inbound chan SipMessage) {
	for {
		b := make([]byte, udpPacketSize)
		_, _, err := connection.ReadFromUDP(b)
		if err != nil {
			log.Println("Error on read: ", err)
			continue
		}
		firstLine, sipHeaders, sipMessageBody, err := DeserializeSipMessage(b, false)
		if err != nil {
			log.Println(err)
		}
		inbound <- SipMessage{FirstLine: firstLine, Headers: sipHeaders, Body: sipMessageBody}
	}
}
