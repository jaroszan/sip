// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"errors"
	"strconv"
	"strings"
)

// REQUEST is used for processing SIP requests
// RESPONSE is used for processing SIP responses
const (
	REQUEST  = "request"
	RESPONSE = "response"
)

type Dialog struct {
	toTag            string
	fromTag          string
	transport        string
	remoteContactUri string
}

var existingDialogs map[string]Dialog

var existingTags map[string]string

func init() {
	existingTags = make(map[string]string)
	existingDialogs = make(map[string]Dialog)
}

// PrepareResponse builds basic structure for a response to an incoming request
func PrepareResponse(sipHeaders map[string]string, code int, reasonPhrase string) string {
	if code > 100 {
		if _, ok := existingTags[sipHeaders["call-id"]]; !ok {
			existingTags[sipHeaders["call-id"]] = sipHeaders["to"] + ";tag=" + GenerateTag()
		}
	}
	var response string
	response += "SIP/2.0 " + strconv.Itoa(code) + " " + reasonPhrase + "\r\n"
	response += "Via: " + sipHeaders["via"] + "\r\n"
	if code == 100 {
		response += "To: " + sipHeaders["to"] + "\r\n"
	} else {
		response += "To: " + existingTags[sipHeaders["call-id"]] + "\r\n"
		//response += "Contact: " + "sip:bob@localhost:5060" + "\r\n"
	}
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

// NewDialog prepares INVITE for initiation of a new dialog
func NewDialog(fromUri string, toUri string, transport string) string {
	callID := GenerateCallID()
	fromTag := GenerateTag()
	existingDialogs[callID] = Dialog{fromTag: fromTag, transport: transport}
	var request string
	request += "INVITE " + toUri + " SIP/2.0" + "\r\n"
	request += "Via: " + "SIP/2.0/" + transport + " " + "localhost:5160;branch=" + GenerateBranch() + "\r\n"
	request += "To: " + toUri + "\r\n"
	request += "From: " + fromUri + ";tag=" + fromTag + "\r\n"
	request += "Call-ID: " + callID + "\r\n"
	request += "CSeq: 1 INVITE" + "\r\n"
	request += "Max-Forwards: 70" + "\r\n"
	request += "Content-Length: 0" + "\r\n"
	request += "Contact: " + fromUri + "\r\n" + "\r\n"
	return request
}

// PrepareInDialogRequest prepares in-dialog requests
//func PrepareInDialogRequest(method string, cseq string, transport string, sipHeaders map[string]string) string {
func PrepareInDialogRequest(method string, cseq string, sipHeaders map[string]string) string {
	tmp := existingDialogs[sipHeaders["call-id"]]
	tmpContact := strings.TrimPrefix(sipHeaders["contact"], "<")
	tmpContact = strings.TrimSuffix(tmpContact, ">")
	tmp.remoteContactUri = tmpContact
	existingDialogs[sipHeaders["call-id"]] = tmp
	var request string
	request += method + " " + existingDialogs[sipHeaders["call-id"]].remoteContactUri + " SIP/2.0" + "\r\n"
	request += "Via: " + "SIP/2.0/" + existingDialogs[sipHeaders["call-id"]].transport + " " + "localhost:5160;branch=" + GenerateBranch() + "\r\n"
	request += "To: " + sipHeaders["to"] + "\r\n"
	request += "From: " + sipHeaders["from"] + "\r\n"
	request += "Call-ID: " + sipHeaders["call-id"] + "\r\n"
	request += "CSeq: " + cseq + " " + method + "\r\n"
	request += "Max-Forwards: 70" + "\r\n"
	request += "Content-Length: 0" + "\r\n" + "\r\n"
	//request += "Contact: sip:bob@localhost:5160" + "\r\n" + "\r\n"
	return request
}
