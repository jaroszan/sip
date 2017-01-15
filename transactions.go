// Copyright 2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"log"
	"strings"
	"sync"
	"time"
	"strconv"
)

const timerAInitialValue = 500
const timerATimeout = 32000

type InviteClientState uint8

//FSM states for Invite Client Transactions
const (
	InviteClientCalling InviteClientState = iota + 1
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

//type transactionList map[string]ClientTransaction

type ActiveClientTransactions struct {
	mtx sync.Mutex
	trm map[string]ClientTransaction
}

func (act *ActiveClientTransactions) NewTransaction(id string, outgoingMessage SipMessage) {
	act.mtx.Lock()
	act.trm[id] = ClientTransaction{currentState: InviteClientCalling, initialRequest: outgoingMessage}
	act.mtx.Unlock()
}

func (act *ActiveClientTransactions) DeleteTransaction(id string) {
	act.mtx.Lock()
	delete(act.trm, id)
	act.mtx.Unlock()
}

func (act *ActiveClientTransactions) GetState(id string) InviteClientState {
	act.mtx.Lock()
	state := act.trm[id].currentState
	act.mtx.Unlock()
	return state
}

func (act *ActiveClientTransactions) UpdateState(id string, newState InviteClientState) {
	act.mtx.Lock()
	tmp := act.trm[id]
	tmp.currentState = newState
	act.trm[id] = tmp
	act.mtx.Unlock()
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

func retransmitFunction(outbound chan SipMessage, sipMessage SipMessage, transactionID string, Act *ActiveClientTransactions) {
	if sipMessage.GetHeaders()["Cseq"] != "1 ACK" {
		transactionTimer := time.Duration(timerAInitialValue)
		timerATransactionTimeout := time.Duration(timerATimeout)
		for {
			time.Sleep(time.Millisecond * transactionTimer)
			if Act.GetState(transactionID) == InviteClientCalling {
				outbound <- sipMessage
				transactionTimer = transactionTimer * 2
				if transactionTimer >= timerATransactionTimeout {
					log.Println("REQUEST TIMEOUT")
					return
				}
			} else {
				return
			}
		}
	}
	return
}

func TransactionLayerHandler(outboundT chan SipMessage, outbound chan SipMessage, inboundT chan SipMessage, inbound chan SipMessage) {
	trl := make(map[string]ClientTransaction)	
	// data structure for holding active transactions with branch parameter as the map key
	Act := &ActiveClientTransactions{trm: trl}
	for {
		select {
		case outgoingMessage := <- outboundT:
			transactionID := getViaBranchValue(outgoingMessage.GetHeaders()["Via"])
			Act.NewTransaction(transactionID, outgoingMessage)
			outbound <- outgoingMessage
			go retransmitFunction(outbound, outgoingMessage, transactionID, Act)
		case incomingMessage := <- inbound:
			transactionID := getViaBranchValue(incomingMessage.GetHeaders()["via"])
			if incomingMessage.Type() == TypeResponse {
				//transactionState := Act.GetState(transactionID)
				statusCode := getStatusCode(incomingMessage.GetFirstLine())
				if statusCode >= 100 && statusCode < 200 {
					Act.UpdateState(transactionID, InviteClientProceeding)
				} else if statusCode >= 200 && statusCode < 300 {
					Act.UpdateState(transactionID, InviteClientTerminated)
				} else {
					Act.UpdateState(transactionID, InviteClientTerminated)
				}
				inboundT <- incomingMessage
				if Act.GetState(transactionID) == InviteClientTerminated {
					Act.DeleteTransaction(transactionID)
				}
			} else {
				//TODO handle incoming transactions
			}
		}
	}
}

func getStatusCode(firstLine string) int {
	firstLineFields := strings.Fields(firstLine)	
	statusCode, _ := strconv.Atoi(firstLineFields[1])
	return statusCode
}
