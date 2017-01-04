// Copyright 2016 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"github.com/jaroszan/sip"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

type InviteClientState uint8

//FSM states for Invite Client Transactions
const (
	InviteClientCalling InviteClientState = iota
	InviteClientProceeding
	InviteClientCompleted
	InviteClientTerminated
)

//ClientTransaction describes client transaction
type ClientTransaction struct {
	currentState   InviteClientState
	initialRequest sip.SipMessage
	//destination    string
	//transport      string
}

//Create mutex to protect existingSessions
var mu = &sync.RWMutex{}

type sessionData struct {
	ReceivedOK uint8
}

var existingSessions map[string]sessionData
var activeTransactions map[string]ClientTransaction

func init() {
	existingSessions = make(map[string]sessionData)
	activeTransactions = make(map[string]ClientTransaction)
}

func getViaBranchValue(viaHeader string) string {
	var transactionID string
	viaParts := strings.Split(viaHeader, ";")
	for _, value := range viaParts {
		if strings.HasPrefix(value, "branch") {
			branchString := strings.Split(value, "=")
			transactionID = branchString[1]
			break
		}
	}
	return transactionID
}

func retransmitFunction(outbound chan sip.SipMessage, sipMessage sip.SipMessage, activeTransactions map[string]ClientTransaction, transactionID string) {
	if sipMessage.Headers["Cseq"] != "1 ACK" {
		for {
			time.Sleep(time.Second * 2)
			mu.Lock()
			currentTransactionState := activeTransactions[transactionID].currentState
			mu.Unlock()
			if currentTransactionState == InviteClientCalling {
				log.Println("Retransmitting")
				log.Println(sipMessage.Headers)
				outbound <- sipMessage
			} else {
				return
			}
		}
	}
	return
}

func outgoingTransactionLayer(outboundTransactions chan sip.SipMessage, outbound chan sip.SipMessage) {
	for newMessage := range outboundTransactions {
		transactionID := getViaBranchValue(newMessage.Headers["Via"])
		mu.Lock()
		activeTransactions[transactionID] = ClientTransaction{currentState: InviteClientCalling, initialRequest: newMessage}
		mu.Unlock()
		outbound <- newMessage
		go retransmitFunction(outbound, newMessage, activeTransactions, transactionID)
	}
}

func incomingTransactionLayer(inboundTransactions chan sip.SipMessage, inbound chan sip.SipMessage) {
	for incomingMessage := range inbound {
		transactionID := getViaBranchValue(incomingMessage.Headers["via"])
		mu.Lock()
		myTmp := activeTransactions
		mu.Unlock()
		if _, ok := myTmp[transactionID]; ok {
			mType, _, err := sip.ParseFirstLine(incomingMessage.FirstLine)
			if err != nil {
				log.Println(err)
			}
			if mType == sip.REQUEST {
				log.Println("Request detected")
				//PROCESS REQUESTS HERE
			} else {
				mu.Lock()
				tmp := activeTransactions[transactionID]
				tmp.currentState = InviteClientProceeding
				activeTransactions[transactionID] = tmp
				mu.Unlock()
				inboundTransactions <- incomingMessage
			}
		} else {
			log.Println("Requests go here!!!")
		}
	}
}

func handleIncomingPacket(inboundTransactions chan sip.SipMessage, outboundTransactions chan sip.SipMessage) {
	for sipMessage := range inboundTransactions {
		sipMessage := sipMessage
		go func() {
			mType, mValue, err := sip.ParseFirstLine(sipMessage.FirstLine)
			if err != nil {
				log.Println(err)
				log.Println("Dropping request")
				runtime.Goexit()
			}

			if mType == sip.REQUEST {
				//PROCESS REQUESTS HERE
			} else if mType == sip.RESPONSE {
				mu.Lock()
				if _, ok := existingSessions[sipMessage.Headers["call-id"]]; !ok {
					existingSessions[sipMessage.Headers["call-id"]] = sessionData{0}
				}
				mu.Unlock()
				if mValue == "200" {
					if sipMessage.Headers["cseq"] == "1 INVITE" {
						mu.Lock()
						isOkReceived := existingSessions[sipMessage.Headers["call-id"]].ReceivedOK
						mu.Unlock()
						if isOkReceived == 0 {
							mu.Lock()
							existingSessions[sipMessage.Headers["call-id"]] = sessionData{1}
							mu.Unlock()
							ackRequest := sip.PrepareInDialogRequest("ACK", "1", sipMessage.Headers)
							outboundTransactions <- ackRequest
							byeRequest := sip.PrepareInDialogRequest("BYE", "2", sipMessage.Headers)
							//go func() {
							time.Sleep(time.Second * 2)
							outboundTransactions <- byeRequest
							//}()
						} else {
							log.Println("Retransmission received")
						}
					} else if sipMessage.Headers["cseq"] == "2 BYE" {
						mu.Lock()
						delete(existingSessions, sipMessage.Headers["call-id"])
						mu.Unlock()
					}
				} else if mValue != "200" {
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
	// outboundTransactions is channel for passing ClientTransactions from TU (handleIncomingPacket) to Transaction Layer (handleTransactions)
	outboundTransactions := make(chan sip.SipMessage)
	// inboundTransactions is channel for passing ClientTransactions from Transaction Layer (handleTransactions) to TU (handleIncomingPacket)
	inboundTransactions := make(chan sip.SipMessage)
	// Initiate TCP connection to remote peer, inbound/outbound are channels are used
	// for receiving and sending messages respectively
	inbound, outbound := sip.StartSIP(localAddr, remoteAddr, transport)
	// Goroutine for processing incoming messages
	go handleIncomingPacket(inboundTransactions, outboundTransactions)
	go incomingTransactionLayer(inboundTransactions, inbound)
	go outgoingTransactionLayer(outboundTransactions, outbound)
	ticker := time.NewTicker(time.Millisecond * 25)
	go func() {
		for _ = range ticker.C {
			// Prepare INVITE
			newRequest := sip.NewDialog("sip:bob@"+localAddr, "sip:alice@"+remoteAddr, transport)
			outboundTransactions <- newRequest
		}
	}()
	time.Sleep(time.Millisecond * 300000)
	ticker.Stop()
	time.Sleep(time.Second * 5)
}

/*
func generateNewTransactionID(viaHeader string, cseqHeader string) string {
	viaBranchParam := getViaBranchValue(sipMessage.Headers["via"])
	cseqMethod := getCseqMethod(sipMessage.Headers["cseq"])
	return viaBranchParam + cseqMethod
}

func getCseqMethod(cseqHeader string) string {
	var cseqMethod string
	cseqParts := strings.Fields(cseqHeader)
	cseqMethod = cseqParts[1]
	return cseqMethod
}*/
