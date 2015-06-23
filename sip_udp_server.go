package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

const (
	REQUEST  = "request"
	RESPONSE = "response"
)

// Prepare a response to incoming request
func prepareResponse(sipHeaders map[string]string, code string, reasonPhrase string) string {
	var response string
	response += "SIP/2.0 " + code + " " + reasonPhrase + "\r\n"
	response += "Via: " + sipHeaders["Via"] + "\r\n"
	response += "To: " + sipHeaders["To"] + "\r\n"
	response += "From: " + sipHeaders["From"] + "\r\n"
	response += "Call-ID: " + sipHeaders["Call-ID"] + "\r\n"
	response += "CSeq: " + sipHeaders["CSeq"] + "\r\n"
	response += "Max-Forwards: " + sipHeaders["Max-Forwards"] + "\r\n"
	return response
}

// Adding header to response, should be called after prepareResponse to add non-mandatory headers
func addHeader(responseHeaders string, newHeaderName string, newHeaderValue string) string {
	responseHeaders += newHeaderName + ": " + newHeaderValue + "\r\n"
	return responseHeaders
}

// Finalizing response and putting it on the wire
func sendResponse(responseHeaders string, addr *net.UDPAddr, connection *net.UDPConn) {
	var payload bytes.Buffer
	responseHeaders += "\r\n"
	payload.WriteString(responseHeaders)
	_, err := connection.WriteToUDP(payload.Bytes(), addr)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

// Parse first line to determine whether incoming message is a response or a request
func parseFirstLine(ruriOrStatusLine string) (string, string, error) {
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
func parseHeaders(headers []string) map[string]string {
	sipHeaders := make(map[string]string)
	for _, value := range headers {
		header := strings.SplitN(value, ":", 2)
		if len(header) != 2 {
			// TODO: processing of SDP body
			break
		}
		sipHeaders[strings.TrimSpace(header[0])] = strings.TrimSpace(header[1])
	}
	return sipHeaders
}

func handleIncomingPacket(buf []byte, addr *net.UDPAddr, packetSize int, connection *net.UDPConn) {
	payload := string(buf[0:packetSize])
	lines := strings.Split(payload, "\n")
	_, mValue, err := parseFirstLine(lines[0])
	//fmt.Println(mType, " ", mValue, " received")
	if err != nil {
		fmt.Println(err)
	}

	sipHeaders := parseHeaders(lines[1:])
	if mValue == "INVITE" {
		outboundTrying := prepareResponse(sipHeaders, "100", "Trying")
		outboundOK := prepareResponse(sipHeaders, "200", "OK")
		outboundOK = addHeader(outboundOK, "Contact", "sip:bob@localhost:5060")
		sendResponse(outboundTrying, addr, connection)
		sendResponse(outboundOK, addr, connection)

	} else if mValue == "BYE" {
		outboundOK := prepareResponse(sipHeaders, "200", "OK")
		sendResponse(outboundOK, addr, connection)
	} else {
	}
}

func listen(connection *net.UDPConn, quit chan struct{}) {
	// 4 Kb seems sufficient for SIP packet
	buffer := make([]byte, 4096)
	n, remoteAddr, err := 0, new(net.UDPAddr), error(nil)
	for err == nil {
		n, remoteAddr, err = connection.ReadFromUDP(buffer)
		packet := buffer
		go handleIncomingPacket(packet, remoteAddr, n, connection)
	}
	fmt.Println("listener failed - ", err)
	quit <- struct{}{}
}

func main() {
	addr := net.UDPAddr{
		Port: 5060,
		IP:   net.IP{127, 0, 0, 1},
	}
	connection, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	quit := make(chan struct{})
	for i := 0; i < runtime.NumCPU(); i++ {
		fmt.Println(runtime.NumCPU(), " listener(s) started as determined by runtime.NumCPU()")
		go listen(connection, quit)
	}
	<-quit
}
