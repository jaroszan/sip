// Copyright 2015 go-sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"errors"
	"strings"
)

const (
	REQUEST  = "request"
	RESPONSE = "response"
)

// Prepare a response to incoming request
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

// Adding header to response, should be called after prepareResponse to add non-mandatory headers
func AddHeader(responseHeaders string, newHeaderName string, newHeaderValue string) string {
	responseHeaders = strings.TrimSpace(responseHeaders)
	responseHeaders += "\r\n" + newHeaderName + ": " + newHeaderValue + "\r\n" + "\r\n"
	return responseHeaders
}

// Parse first line to determine whether incoming message is a response or a request
func ParseFirstLine(ruriOrStatusLine string) (string, string, error) {
	firstLine := strings.Fields(ruriOrStatusLine)
	if firstLine[2] == "SIP/2.0" {
		return REQUEST, firstLine[0], nil
	} else if firstLine[0] == "SIP/2.0" {
		return RESPONSE, firstLine[1], nil
	} else {
		return "", "", errors.New("Incoming datagram doesn't look like a SIP message: " + ruriOrStatusLine)
	}
}

// Parsing headers, so far SDP is not processed in any way
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
