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

func makeResponse(sipHeaders map[string]string, code string, reasonPhrase string, method string) []byte {
	var response bytes.Buffer
	response.WriteString("SIP/2.0 " + code + " " + reasonPhrase + "\r\n")
	response.WriteString("Via: " + sipHeaders["Via"] + "\r\n")
	response.WriteString("To: " + sipHeaders["To"] + "\r\n")
	response.WriteString("From: " + sipHeaders["From"] + "\r\n")
	response.WriteString("Call-ID: " + sipHeaders["Call-ID"] + "\r\n")
	response.WriteString("CSeq: " + sipHeaders["CSeq"] + "\r\n")
	response.WriteString("Content-Length: 0\r\n")
	if method != "BYE" && code != "100" {
		response.WriteString("Contact: sip:bob@localhost:5060\r\n")
	}
	response.WriteString("Max-Forwards: " + sipHeaders["Max-Forwards"] + "\r\n" + "\r\n")
	return response.Bytes()
}

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

//func parseHeaders(

func handleIncomingPacket(buf []byte, addr *net.UDPAddr, packetSize int, connection *net.UDPConn) {
	sipHeaders := make(map[string]string)
	payload := string(buf[0:packetSize])
	lines := strings.Split(payload, "\n")
	mType, mValue, err := parseFirstLine(lines[0])
	fmt.Println(mType, " ", mValue, " received")
	if err != nil {
		fmt.Println(err)
	}

	for _, value := range lines[1:] {
		header := strings.SplitN(value, ":", 2)
		// Ignoring body after headers are processed
		if len(header) != 2 {
			break
		}
		sipHeaders[strings.TrimSpace(header[0])] = strings.TrimSpace(header[1])
	}

	if mValue == "INVITE" {
		outboundTrying := makeResponse(sipHeaders, "100", "Trying", mValue)
		outboundOK := makeResponse(sipHeaders, "200", "", mValue)
		_, err := connection.WriteToUDP(outboundTrying, addr)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}
		_, err = connection.WriteToUDP(outboundOK, addr)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}
	} else if mValue == "BYE" {
		outboundOK := makeResponse(sipHeaders, "200", "OK", mValue)
		_, err := connection.WriteToUDP(outboundOK, addr)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(0)
		}
	} else {
	}
}

func listen(connection *net.UDPConn, quit chan struct{}) {
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
		fmt.Println(runtime.NumCPU(), " processor(s) started")
		go listen(connection, quit)
	}
	<-quit
}
