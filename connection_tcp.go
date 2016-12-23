// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bufio"
	"bytes"
	"log"
	"net"
)

// StartTCPClient prepares TCP connection and returns inbound/outbound channels
func StartTCPClient(lAddr string, rAddr string) (chan []byte, chan []byte, *net.TCPConn) {
	//var wg sync.WaitGroup
	//wg.Add(2)
	//Resolve local address
	localTcpAddr, err := net.ResolveTCPAddr("tcp4", lAddr)
	CheckConnError(err)

	//Resolve remote address
	remoteTcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:5060")
	CheckConnError(err)

	//Establish connection to remote address
	conn, err := net.DialTCP("tcp", localTcpAddr, remoteTcpAddr)
	CheckConnError(err)

	// Outbound channel uses connection to send messages
	outbound := make(chan []byte)
	// Inbound channel passes received message to handleIncomingPacket function
	inbound := make(chan []byte)

	// Goroutine for receiving messages and passing them to handleIncomingPacket function
	go recvTCP(conn, inbound)
	// Goroutine for sending messages
	go sendTCP(conn, outbound)

	return inbound, outbound, conn

}

func sendTCP(connection *net.TCPConn, outbound chan []byte) {
	for message := range outbound {
		_, err := connection.Write(message)
		if err != nil {
			log.Println("Error on write: ", err)
			continue
		}
	}
}

func recvTCP(connection *net.TCPConn, inbound chan []byte) {
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
