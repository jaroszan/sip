// Copyright 2016 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"github.com/jaroszan/sip"
	"log"
	"runtime"
	"sync"
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

func handleIncomingPacket(inbound chan sip.SipMessage, outbound chan sip.SipMessage, wg *sync.WaitGroup) {
	for sipMessage := range inbound {
		sipMessage := sipMessage
		go func() {
			mType, mValue, err := sip.ParseFirstLine(sipMessage.FirstLine)
			if err != nil {
				log.Println(err)
				log.Println("Dropping request")
				runtime.Goexit()
			}

			if mType == sip.REQUEST {
				if mValue == "INVITE" {
					outboundTrying := sip.PrepareResponse(sipMessage.Headers, 100, "Trying")
					outbound180 := sip.PrepareResponse(sipMessage.Headers, 180, "Ringing")
					outbound180 = sip.AddHeader(outbound180, "Contact", "sip:bob@localhost:5060")
					outboundOK := sip.PrepareResponse(sipMessage.Headers, 200, "OK")
					outboundOK = sip.AddHeader(outboundOK, "Contact", "sip:alice@localhost:5060")
					outbound <- outboundTrying
					outbound <- outbound180
					outbound <- outboundOK
				} else if mValue == "BYE" {
					outboundOK := sip.PrepareResponse(sipMessage.Headers, 200, "OK")
					outbound <- outboundOK
				} else {
					log.Println(mValue + " received")
				}
			} else if mType == sip.RESPONSE {
				//PROCESS RESPONSES HERE
			}
		}()
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	// Define local and remote peer
	localAddr := "localhost:5160"
	remoteAddr := "localhost:5060"
	// Define protocol to be used, either TCP or UDP is a valid choice
	transport := "UDP"
	// Initiate TCP connection to remote peer, inbound/outbound are channels are used
	// for receiving and sending messages respectively
	inbound, outbound := sip.StartSIP(localAddr, remoteAddr, transport)
	// Goroutine for processing incoming messages
	go handleIncomingPacket(inbound, outbound, &wg)
	wg.Wait()
}
