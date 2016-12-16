// Copyright 2015 sip authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package sip

import (
	"bytes"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const branchMagicCookie = "z9hG4bK"

// GenerateRandom generates random strings for tags/call-ids/branches
func GenerateRandom(charsToDouble int) string {
	var buf bytes.Buffer
	for i := 0; i < charsToDouble; i++ {
		buf.WriteByte(byte(randInt(65, 90)))
		buf.WriteByte(byte(randInt(97, 122)))
	}

	return string(buf.Bytes())
}

// Helper for generating random number between two given numbers
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// GenerateBranch generates branch for via header
func GenerateBranch() string {
	randomPart := GenerateRandom(4)
	return branchMagicCookie + randomPart
}
