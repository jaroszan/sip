// Copyright 2016 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"strings"
)

func StartSIP(lAddr string, rAddr string, transport string) (chan SipMessage, chan SipMessage) {
	if strings.ToLower(transport) == "udp" {
		return StartUDP(lAddr, rAddr)
	}
	return StartTCPClient(lAddr, rAddr)
}
