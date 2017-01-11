// Copyright 2016-2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bytes"
	"strings"
	"strconv"
)

// DeserializeSipMessage transforms bytes comprising a sip message into SIP message
func DeserializeSipMessage(payload []byte, streamed bool) (SipMessage, error) {
	message := string(payload)
	var lines []string
	var messageParts []string
	// if not streamed it means that we are processing UDP datagram and need to separate body of the message
	if !streamed {
		messageParts = strings.Split(message, "\r\n\r\n")
		lines = strings.Split(messageParts[0], "\n")
	} else {
		// if streamed parsing start line and headers, body if present will be attached to SipMessage later
		// on connection level
		lines = strings.Split(message, "\n")
	}
	rType, firstLine, err := ParseFirstLine(lines[0])
	if err != nil {
		return nil, err
	}
	
	sipHeaders := parseHeaders(lines[1:])
	// if not streamed check if body is present
	if !streamed {
		if len(messageParts[1]) == 2 {
			if rType == REQUEST {
				return Request{Method: firstLine[0], RUri: firstLine[1], SipVersion: firstLine[2], Headers: sipHeaders, Body: messageParts[1]}, nil
			}
			statusCode, _ := strconv.Atoi(firstLine[1])
			return Response{SipVersion: firstLine[0], StatusCode: statusCode, ReasonPhrase: firstLine[2], Headers: sipHeaders, Body: messageParts[1]}, nil
		}
	}
	if rType == REQUEST {
		return Request{Method: firstLine[0], RUri: firstLine[1], SipVersion: firstLine[2], Headers: sipHeaders, Body: ""}, nil
	}
	statusCode, _ := strconv.Atoi(firstLine[1])
	return Response{SipVersion: firstLine[0], StatusCode: statusCode, ReasonPhrase: firstLine[2], Headers: sipHeaders, Body: ""}, nil

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
