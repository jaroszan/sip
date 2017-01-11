// Copyright 2017 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

type MessageType uint8

const (
	TypeRequest MessageType = iota
	TypeResponse
)

type SipMessage interface {
	Type() MessageType
	GetBody() string
	GetHeaders() map[string]string
	GetFirstLine() string
	SetBody(string)
	AddHeader(string, string)
	Serialize() []byte
}
