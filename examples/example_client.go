// Copyright 2016 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"github.com/jaroszan/sip"
	"log"
	"strings"
	"sync"
	"time"
)

//Create mutex to protect existingSessions
var mu = &sync.RWMutex{}

type sessionData struct {
	ReceivedOK uint8
}

var existingSessions map[string]sessionData

func init() {
	existingSessions = make(map[string]sessionData)
}

func handleIncomingPacket(recvChannel chan sip.SipMessage, sendChannel chan sip.SipMessage) {
	for sipMessage := range recvChannel {
		sipMessage := sipMessage
		go func() {
			messageType := sipMessage.Type()
			if messageType == sip.TypeRequest {
				//PROCESS REQUESTS HERE
			} else if messageType == sip.TypeResponse {
				incomingMessageHeaders := sipMessage.GetHeaders()
				mu.Lock()
				if _, ok := existingSessions[incomingMessageHeaders["call-id"]]; !ok {
					existingSessions[incomingMessageHeaders["call-id"]] = sessionData{0}
				}
				mu.Unlock()
				startLine := strings.Fields(sipMessage.GetFirstLine())
				if startLine[1] == "200" {
					if incomingMessageHeaders["cseq"] == "1 INVITE" {
						mu.RLock()
						isOkReceived := existingSessions[incomingMessageHeaders["call-id"]].ReceivedOK
						mu.RUnlock()
						if isOkReceived == 0 {
							mu.Lock()
							existingSessions[incomingMessageHeaders["call-id"]] = sessionData{1}
							mu.Unlock()
							ackRequest := sip.PrepareInDialogRequest("ACK", "1", incomingMessageHeaders)
							sendChannel <- ackRequest
							byeRequest := sip.PrepareInDialogRequest("BYE", "2", incomingMessageHeaders)
							//go func() {
							time.Sleep(time.Second * 2)
							sendChannel <- byeRequest
							//}()
						} else {
							log.Println("Retransmission received")
						}
					} else if sipMessage.GetHeaders()["cseq"] == "2 BYE" {
						mu.Lock()
						delete(existingSessions, incomingMessageHeaders["call-id"])
						mu.Unlock()
					}
				} else if startLine[1] != "200" {
					//log.Println("Provisional response received: " + mValue)
				} else {
					//log.Println("Response received: " + mValue)
				}
			}
		}()
	}
}

func main() {
	// Define local and remote peer
	localAddr := "localhost:5160"
	remoteAddr := "localhost:5060"
	// Define protocol to be used, either TCP or UDP is a valid choice
	transport := "UDP"
	// Initiate TCP connection to remote peer, inbound/outbound are channels are used
	// for receiving and sending messages respectively
	recvChannel, sendChannel := sip.StartSIP(localAddr, remoteAddr, transport)
	// Goroutine for processing incoming messages
	go handleIncomingPacket(recvChannel, sendChannel)
	ticker := time.NewTicker(time.Millisecond * 20)
	go func() {
		for _ = range ticker.C {
			// Prepare INVITE
			newRequest := sip.NewDialog("sip:bob@"+localAddr, "sip:alice@"+remoteAddr, transport)
			sendChannel <- newRequest
		}
	}()
	time.Sleep(time.Millisecond * 30)
	ticker.Stop()
	time.Sleep(time.Second * 10)
}
