package main

import (
	"fmt"
	"github.com/jaroszan/sip"
	"net"
	"strings"
	"sync"
	"log"
)

type Packet struct {
	addr *net.UDPAddr
	data []byte
}

const UDP_PACKET_SIZE = 2048

func handleIncomingPacket(inbound chan Packet, outbound chan Packet, wg *sync.WaitGroup) {
	defer wg.Done()
	for packet := range inbound {
		payload := string(packet.data)
		lines := strings.Split(payload, "\n")
		_, mValue, err := sip.ParseFirstLine(lines[0])

		if err != nil {
			fmt.Println(err)
		}

		sipHeaders := sip.ParseHeaders(lines[1:])
		if mValue == "INVITE" {
			outboundTrying := sip.PrepareResponse(sipHeaders, "100", "Trying")
			outboundOK := sip.PrepareResponse(sipHeaders, "200", "OK")
			outboundOK = sip.AddHeader(outboundOK, "Contact", "sip:bob@localhost:5060")
			outbound <- Packet{packet.addr, []byte(outboundTrying)}
			outbound <- Packet{packet.addr, []byte(outboundOK)}

		} else if mValue == "BYE" {
			outboundOK := sip.PrepareResponse(sipHeaders, "200", "OK")
			outbound <- Packet{packet.addr, []byte(outboundOK)}
		} else {
		}
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(3)

	//rand.Seed(time.Now().UTC().UnixNano())
	connection := sip.Start("127.0.0.1:5060")

	// Outbound channel uses connection to send outgoing datagrams
	outbound := make(chan Packet)
	// Inbound channel passes received datagrams to handleIncomingPacket function
	inbound := make(chan Packet)

	// Goroutine for passing incoming datagrams to handleIncomingPacket function
	go recv(connection, inbound, &wg)
	// Goroutine for processing incoming datagrams
	go handleIncomingPacket(inbound, outbound, &wg)
	// Goroutine for sending outgoing datagrams
	go send(connection, outbound, &wg)
	wg.Wait()

}

func send(connection *net.UDPConn, outbound chan Packet, wg *sync.WaitGroup) {
	defer wg.Done()
	for packet := range outbound {
		_, err := connection.WriteToUDP(packet.data, packet.addr)
		if err != nil {
			log.Println("Error on write: ", err)
			continue
		}
	}
}

func recv(connection *net.UDPConn, inbound chan Packet, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		b := make([]byte, UDP_PACKET_SIZE)
		n, addr, err := connection.ReadFromUDP(b)
		if err != nil {
			log.Println("Error on read: ", err)
			continue
		}
		inbound <- Packet{addr, b[:n]}
	}
}