// Copyright 2016 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bytes"
	"strings"
)

type SipMessage struct {
	FirstLine string
	Headers   map[string]string
	Body      string
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

// DeserializeSipMessage transforms bytes comprising a sip message into SIP message
func DeserializeSipMessage(payload []byte, streamed bool) (string, map[string]string, string, error) {
	message := string(payload)
	var lines []string
	var messageParts []string
	// if not streamed it means that we are processing UDP datagram and need to separate body of the message
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
	// if not streamed check if body is present
	if !streamed {
		if len(messageParts[1]) == 2 {
			return lines[0], sipHeaders, messageParts[1], nil
		}
	}
	return lines[0], sipHeaders, "", nil
}

func parseHeaders(headers []string) map[string]string {
	sipHeaders := make(map[string]string)
	for _, value := range headers {
		header := strings.SplitN(value, ":", 2)
		// If empty line is encountered it means headers' processing is finished
		if len(header) != 2 {
			break
		}
		sipHeaders[strings.TrimSpace(strings.ToLower(header[0]))] = strings.TrimSpace(header[1])
	}
	return sipHeaders
}
