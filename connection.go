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
