// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

// Packet stores UDP connection data and payload
type Packet struct {
	addr *net.UDPAddr
	data []byte
}

// MaxPayload sets the largest accepted UDP packet
const MaxPayload = 2048

// StartUDP prepares and returns UDP connection
func StartUDP(address string) (*net.UDPConn, *net.UDPAddr) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	return connection, addr
}

// StartTCPClient prepares TCP connection and returns inbound/outbound channels
func StartTCPClient(lAddr string, rAddr string) (chan []byte, chan []byte) {
	var wg sync.WaitGroup
	wg.Add(2)
	//Resolve local address
	localTcpAddr, err := net.ResolveTCPAddr("tcp4", lAddr)
	catchError(err)

	//Resolve remote address
	remoteTcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:5060")
	catchError(err)

	//Establish connection to remote address
	conn, err := net.DialTCP("tcp", localTcpAddr, remoteTcpAddr)
	catchError(err)

	// Outbound channel uses connection to send messages
	outbound := make(chan []byte)
	// Inbound channel passes received message to handleIncomingPacket function
	inbound := make(chan []byte)

	// Goroutine for receiving messages and passing them to handleIncomingPacket function
	go recvTCP(conn, inbound)
	// Goroutine for sending messages
	go sendTCP(conn, outbound)

	return inbound, outbound

}

func sendTCP(connection *net.TCPConn, outbound chan []byte) {
	//defer wg.Done()
	for message := range outbound {
		_, err := connection.Write(message)
		if err != nil {
			log.Println("Error on write: ", err)
			continue
		}
	}
}

func recvTCP(connection *net.TCPConn, inbound chan []byte) {
	//defer wg.Done()
	for {
		scanner := bufio.NewScanner(connection)
		onSipDelimiter := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			delim := []byte{'\r', '\n', '\r', '\n'}
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := bytes.Index(data, delim); i > 0 {
				return i + len(delim), data[0:i], nil
			}
			if atEOF {
				return len(data), data, nil
			}
			return 0, nil, nil
		}
		scanner.Split(onSipDelimiter)
		for scanner.Scan() {
			inbound <- scanner.Bytes()
		}
	}
}

func catchError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
