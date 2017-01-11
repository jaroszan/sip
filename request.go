// Copyright 2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bytes"
)

type Request struct {
	Method     string
	RUri       string
	SipVersion string
	Headers    map[string]string
	Body       string
}

func (req Request) Type() MessageType {
	return TypeRequest
}

func (req Request) GetBody() string {
	return req.Body
}

func (req Request) GetHeaders() map[string]string {
	return req.Headers
}

func (req Request) GetFirstLine() string {
	return req.Method + " " + req.RUri + " " + req.SipVersion
}

func (req Request) SetBody(b string) {
	req.Body = b
}

func (req Request) AddHeader(name string, value string) {
	req.Headers[name] = value
}

func (req Request) Serialize() []byte {
	var serializedMessage bytes.Buffer
	serializedMessage.WriteString(req.Method)
	serializedMessage.WriteString(" ")
	serializedMessage.WriteString(req.RUri)
	serializedMessage.WriteString(" ")
	serializedMessage.WriteString(req.SipVersion)
	serializedMessage.WriteString("\r\n")
	for name, value := range req.Headers {
		serializedMessage.WriteString(name)
		serializedMessage.WriteString(": ")
		serializedMessage.WriteString(value)
		serializedMessage.WriteString("\r\n")
	}
	serializedMessage.WriteString("\r\n")
	if len(req.Body) > 0 {
		serializedMessage.WriteString(req.Body)
		serializedMessage.WriteString("\r\n")
	}
	return serializedMessage.Bytes()
}
