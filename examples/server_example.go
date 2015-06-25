package main

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
	"github.com/jaroszan/sip"
)



func handleIncomingPacket(buf []byte, addr *net.UDPAddr, packetSize int, connection *net.UDPConn) {
	payload := string(buf[0:packetSize])
	lines := strings.Split(payload, "\n")
	_, mValue, err := sip.ParseFirstLine(lines[0])
	//fmt.Println(mType, " ", mValue, " received")
	if err != nil {
		fmt.Println(err)
	}

	sipHeaders := sip.ParseHeaders(lines[1:])
	if mValue == "INVITE" {
		outboundTrying := sip.PrepareResponse(sipHeaders, "100", "Trying")
		outboundOK := sip.PrepareResponse(sipHeaders, "200", "OK")
		outboundOK = sip.AddHeader(outboundOK, "Contact", "sip:bob@localhost:5060")
		sip.SendResponse(outboundTrying, addr, connection)
		sip.SendResponse(outboundOK, addr, connection)

	} else if mValue == "BYE" {
		outboundOK := sip.PrepareResponse(sipHeaders, "200", "OK")
		sip.SendResponse(outboundOK, addr, connection)
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
	rand.Seed(time.Now().UTC().UnixNano())
	connection := sip.Start("127.0.0.1:5060")
	quit := make(chan struct{})
	go listen(connection, quit)
	<-quit
}
