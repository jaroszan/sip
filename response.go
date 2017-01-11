// Copyright 2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"strconv"
	"bytes"
)

type Response struct {
	SipVersion   string
	StatusCode   int
	ReasonPhrase string
	Headers      map[string]string
	Body         string
}

func (res Response) Type() MessageType {
	return TypeResponse
}

func (res Response) GetBody() string {
	return res.Body
}

func (res Response) GetHeaders() map[string]string {
	return res.Headers
}

func (res Response) GetFirstLine() string {
	stringifiedStatusCode := strconv.Itoa(res.StatusCode)
	return res.SipVersion + " " + stringifiedStatusCode + " " + res.ReasonPhrase
}

func (res Response) SetBody(b string) {
	res.Body = b
}

func (res Response) AddHeader(name string, value string) {
	res.Headers[name] = value
}

func (res Response) Serialize() []byte {
	var serializedMessage bytes.Buffer
	serializedMessage.WriteString(res.SipVersion)
	serializedMessage.WriteString(" ")
	stringifiedStatusCode := strconv.Itoa(res.StatusCode)
	serializedMessage.WriteString(stringifiedStatusCode)
	serializedMessage.WriteString(" ")
	serializedMessage.WriteString(res.ReasonPhrase)
	serializedMessage.WriteString("\r\n")
	for name, value := range res.Headers {
		serializedMessage.WriteString(name)
		serializedMessage.WriteString(": ")
		serializedMessage.WriteString(value)
		serializedMessage.WriteString("\r\n")
	}
	serializedMessage.WriteString("\r\n")
	if len(res.Body) > 0 {
		serializedMessage.WriteString(res.Body)
		serializedMessage.WriteString("\r\n")
	}
	return serializedMessage.Bytes()

}
