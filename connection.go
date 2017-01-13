// Copyright 2016-2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"strings"
)

func StartSIP(lAddr string, rAddr string, transport string) (chan SipMessage, chan SipMessage) {
	// prepare and start goroutines and channels responsible for handling
	// incoming/outgoing traffic
	inbound := make(chan SipMessage)
	outbound := make(chan SipMessage)

	if strings.ToLower(transport) == "udp" {
		inbound, outbound = StartUDP(lAddr, rAddr)
	} else {
		inbound, outbound = StartTCPClient(lAddr, rAddr)
	}

	// sendChannel is channel for passing ClientTransactions from Transaction User to Transaction Layer
	sendChannel := make(chan SipMessage)
	// recvChannel is channel for passing ClientTransactions from Transaction Layer to Transaction User
	recvChannel := make(chan SipMessage)

	// prepare and start goroutines and channels responsible for performing
	// transaction layer functions
	go IncomingTransactionLayerHandler(recvChannel, inbound)
	go OutgoingTransactionLayerHandler(sendChannel, outbound)
	return recvChannel, sendChannel
}
