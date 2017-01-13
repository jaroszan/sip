// Copyright 2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"log"
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
	initialRequest SipMessage
	//destination    string
	//transport      string
}

var activeTransactions map[string]ClientTransaction

//Create mutex to protect activeTransactions
var mx = &sync.RWMutex{}

func init() {
	// data structure for holding active transactions with branch parameter as the map key
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

func retransmitFunction(outbound chan SipMessage, sipMessage SipMessage, transactionID string) {
	if sipMessage.GetHeaders()["Cseq"] != "1 ACK" {
		for {
			time.Sleep(time.Second * 2)
			mx.RLock()
			currentTransactionState := activeTransactions[transactionID].currentState
			mx.RUnlock()
			if currentTransactionState == InviteClientCalling {
				log.Println("Retransmitting")
				log.Println(sipMessage.GetHeaders())
				outbound <- sipMessage
			} else {
				return
			}
		}
	}
	return
}

func OutgoingTransactionLayerHandler(outboundTransactions chan SipMessage, outbound chan SipMessage) {
	for outgoingMessage := range outboundTransactions {
		transactionID := getViaBranchValue(outgoingMessage.GetHeaders()["Via"])
		mx.Lock()
		activeTransactions[transactionID] = ClientTransaction{currentState: InviteClientCalling, initialRequest: outgoingMessage}
		mx.Unlock()
		outbound <- outgoingMessage
		go retransmitFunction(outbound, outgoingMessage, transactionID)
	}
}

func IncomingTransactionLayerHandler(inboundTransactions chan SipMessage, inbound chan SipMessage) {
	for incomingMessage := range inbound {
		transactionID := getViaBranchValue(incomingMessage.GetHeaders()["via"])
		mx.RLock()
		myTmp := activeTransactions
		mx.RUnlock()
		if _, ok := myTmp[transactionID]; ok {
			if incomingMessage.Type() == TypeRequest {
				//TODO request processing
			} else {
				mx.Lock()
				tmp := activeTransactions[transactionID]
				tmp.currentState = InviteClientProceeding
				activeTransactions[transactionID] = tmp
				mx.Unlock()
				inboundTransactions <- incomingMessage
			}
		} else {
			//TODO new transaction handling
		}
	}
}
