// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strconv"
)

// StartTCPClient prepares TCP connection and returns inbound/outbound channels
func StartTCPClient(lAddr string, rAddr string) (chan SipMessage, chan SipMessage) {
	//Resolve local address
	localTcpAddr, err := net.ResolveTCPAddr("tcp4", lAddr)
	CheckConnError(err)

	//Resolve remote address
	remoteTcpAddr, err := net.ResolveTCPAddr("tcp4", rAddr)
	CheckConnError(err)

	//Establish connection to remote address
	conn, err := net.DialTCP("tcp", localTcpAddr, remoteTcpAddr)
	CheckConnError(err)

	// Outbound channel uses connection to send messages
	outbound := make(chan SipMessage)
	// Inbound channel passes received messages to processing goroutine
	inbound := make(chan SipMessage)

	// Goroutine for receiving messages and passing them to processing function
	go recvTCP(conn, inbound)
	// Goroutine for sending messages
	go sendTCP(conn, outbound)

	return inbound, outbound

}

func sendTCP(connection *net.TCPConn, outbound chan SipMessage) {
	for message := range outbound {
		_, err := connection.Write(SerializeSipMessage(message))
		if err != nil {
			log.Println("Error on write: ", err)
			continue
		}
	}
}

func recvTCP(connection *net.TCPConn, inbound chan SipMessage) {
	reader := bufio.NewReader(connection)
	for {
		var buffer bytes.Buffer
		for {
			chunk, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
			}
			buffer.WriteString(chunk)
			if len(chunk) == 2 {
				break
			}
		}
		firstLine, sipHeaders, _, _ := ParseIncomingMessage(buffer.Bytes(), true)
		contentLength, err := strconv.Atoi(sipHeaders["content-length"])
		if err != nil {
			log.Println("Content-Length value cannot be converted to number, setting it to 0")
			contentLength = 0
		}
		if contentLength < 0 {
			log.Println("Content-Length value is less than 0, setting it to 0")
			contentLength = 0
		}
		sipMessageBody := []byte{}
		if contentLength != 0 {
			sipMessageBody = make([]byte, contentLength, contentLength)
			_, err := reader.Read(sipMessageBody)
			if err != nil {
				log.Println(err)
			}
		}
		inbound <- SipMessage{FirstLine: firstLine, Headers: sipHeaders, Body: string(sipMessageBody)}
	}
}
