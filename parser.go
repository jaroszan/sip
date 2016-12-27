// Copyright 2015-2016 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bytes"
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

type SipMessage struct {
	FirstLine string
	Headers   map[string]string
	Body      string
}

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
func PrepareResponse(sipHeaders map[string]string, code int, reasonPhrase string) SipMessage {
	var toHeader string
	if code > 100 {
		if _, ok := existingTags[sipHeaders["call-id"]]; !ok {
			existingTags[sipHeaders["call-id"]] = sipHeaders["to"] + ";tag=" + GenerateTag()
		}
		toHeader = existingTags[sipHeaders["call-id"]]
	}
	if code == 100 {
		toHeader = sipHeaders["to"]
	}
	firstLine := "SIP/2.0 " + strconv.Itoa(code) + " " + reasonPhrase
	sipBody := []byte{}
	headers := map[string]string{
		"Via":            sipHeaders["via"],
		"Call-ID":        sipHeaders["call-id"],
		"To":             toHeader,
		"From":           sipHeaders["from"],
		"Cseq":           sipHeaders["cseq"],
		"Content-Length": strconv.Itoa(len(sipBody)),
	}
	return SipMessage{FirstLine: firstLine, Headers: headers, Body: string(sipBody)}
}

// AddHeader adds user-specified non-mandatory header to a message
func AddHeader(sipMessage SipMessage, newHeaderName string, newHeaderValue string) SipMessage {
	sipMessage.Headers[newHeaderName] = newHeaderValue
	return sipMessage
}

//ParseIncomingMessage is used to parse incoming message
//streamed parameter determines transport type
func ParseIncomingMessage(payload []byte, streamed bool) (string, map[string]string, string, error) {
	message := string(payload)
	var lines []string
	var messageParts []string
	if !streamed {
		messageParts = strings.Split(message, "\r\n\r\n")
		lines = strings.Split(messageParts[0], "\n")
	} else {
		lines = strings.Split(message, "\n")
	}
	_, _, err := ParseFirstLine(lines[0])
	if err != nil {
		return "", nil, "", err
	}
	sipHeaders := parseHeaders(lines[1:])
	if !streamed {
		if len(messageParts[1]) == 2 {
			return lines[0], sipHeaders, messageParts[1], nil
		}
	}
	return lines[0], sipHeaders, "", nil
}

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

func parseHeaders(headers []string) map[string]string {
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
func NewDialog(fromUri string, toUri string, transport string) SipMessage {
	callID := GenerateCallID()
	fromTag := GenerateTag()
	newBranch := GenerateBranch()
	existingDialogs[callID] = Dialog{fromTag: fromTag, transport: transport}
	firstLine := "INVITE " + toUri + " SIP/2.0"
	headers := map[string]string{
		"Via":            "SIP/2.0/" + transport + " " + "localhost:5160;branch=" + newBranch,
		"To":             toUri,
		"From":           fromUri + ";tag=" + fromTag,
		"Call-ID":        callID,
		"Cseq":           "1 INVITE",
		"Max-Forwards":   "70",
		"Content-Length": "0",
		"Contact":        fromUri,
	}
	sdpBody := ""
	return SipMessage{FirstLine: firstLine, Headers: headers, Body: sdpBody}
}

// PrepareInDialogRequest prepares in-dialog requests
//func PrepareInDialogRequest(method string, cseq string, transport string, sipHeaders map[string]string) string {
func PrepareInDialogRequest(method string, cseq string, sipHeaders map[string]string) SipMessage {
	tmp := existingDialogs[sipHeaders["call-id"]]
	tmpContact := strings.TrimPrefix(sipHeaders["contact"], "<")
	tmpContact = strings.TrimSuffix(tmpContact, ">")
	tmp.remoteContactUri = tmpContact
	newBranch := GenerateBranch()
	existingDialogs[sipHeaders["call-id"]] = tmp
	firstLine := method + " " + existingDialogs[sipHeaders["call-id"]].remoteContactUri + " SIP/2.0"
	headers := map[string]string{
		"Via":            "SIP/2.0/" + existingDialogs[sipHeaders["call-id"]].transport + " " + "localhost:5160;branch=" + newBranch,
		"To":             sipHeaders["to"],
		"From":           sipHeaders["from"],
		"Call-ID":        sipHeaders["call-id"],
		"Cseq":           cseq + " " + method,
		"Max-Forwards":   "70",
		"Content-Length": "0",
	}
	sdpBody := ""
	return SipMessage{FirstLine: firstLine, Headers: headers, Body: sdpBody}
}

//TODO put mandatory headers first
// SerializeSipMessage serializes SipMessage structure according to SIP message format (RFC 3261)
func SerializeSipMessage(sipMessage SipMessage) []byte {
	var serializedMessage bytes.Buffer
	serializedMessage.WriteString(sipMessage.FirstLine)
	serializedMessage.WriteString("\r\n")
	for name, value := range sipMessage.Headers {
		serializedMessage.WriteString(name)
		serializedMessage.WriteString(": ")
		serializedMessage.WriteString(value)
		serializedMessage.WriteString("\r\n")
	}
	serializedMessage.WriteString("\r\n")
	if len(sipMessage.Body) > 0 {
		serializedMessage.WriteString(sipMessage.Body)
		serializedMessage.WriteString("\r\n")
	}
	return serializedMessage.Bytes()
}
