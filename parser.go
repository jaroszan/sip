// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"errors"
	"strings"
)

// REQUEST is used for processing SIP requests
// RESPONSE is used for processing SIP responses
const (
	REQUEST  = "request"
	RESPONSE = "response"
)

// PrepareResponse builds basic structure for a response to an incoming request
func PrepareResponse(sipHeaders map[string]string, code string, reasonPhrase string) string {
	var response string
	response += "SIP/2.0 " + code + " " + reasonPhrase + "\r\n"
	response += "Via: " + sipHeaders["via"] + "\r\n"
	response += "To: " + sipHeaders["to"] + "\r\n"
	response += "From: " + sipHeaders["from"] + "\r\n"
	response += "Call-ID: " + sipHeaders["call-id"] + "\r\n"
	response += "CSeq: " + sipHeaders["cseq"] + "\r\n"
	response += "Max-Forwards: " + sipHeaders["max-forwards"] + "\r\n" + "\r\n"
	return response
}

// AddHeader adds user-specified non-mandatory header to a message
func AddHeader(responseHeaders string, newHeaderName string, newHeaderValue string) string {
	responseHeaders = strings.TrimSpace(responseHeaders)
	responseHeaders += "\r\n" + newHeaderName + ": " + newHeaderValue + "\r\n" + "\r\n"
	return responseHeaders
}

// ParseFirstLine is used to determine type of the incoming message
func ParseFirstLine(incomingFirstLine string) (string, string, error) {
	firstLine := strings.Fields(incomingFirstLine)
	if firstLine[2] == "SIP/2.0" {
		return REQUEST, firstLine[0], nil
	} else if firstLine[0] == "SIP/2.0" {
		return RESPONSE, firstLine[1], nil
	} else {
		return "", "", errors.New("Incoming datagram doesn't look like a SIP message: " + incomingFirstLine)
	}
}

// ParseHeaders parses headers in the incoming message
func ParseHeaders(headers []string) map[string]string {
	sipHeaders := make(map[string]string)
	for _, value := range headers {
		header := strings.SplitN(value, ":", 2)
		if len(header) != 2 {
			// TODO: processing of SDP body
			break
		}
		sipHeaders[strings.TrimSpace(strings.ToLower(header[0]))] = strings.TrimSpace(header[1])
	}
	return sipHeaders
}

//MakeRequest prepares request
func MakeRequest(method string) string {
	var request string
	request += "INVITE " + "alice@localhost:5060 " + "SIP/2.0" + "\r\n"
	request += "Via: " + "SIP/2.0/UDP" + "localhost:5160;branch=" + GenerateBranch() + "\r\n"
	request += "To: alice@localhost:5060;tag=" + GenerateRandom(4) + "\r\n"
	request += "From: bob@localhost:5160;tag=" + GenerateRandom(4) + "\r\n"
	request += "Call-ID: " + GenerateRandom(4) + "\r\n"
	request += "CSeq: INVITE 1" + "\r\n"
	request += "Max-Forwards: 70" + "\r\n"
	request += "Contact: bob@localhost:5160" + "\r\n" + "\r\n"

	return request
}
