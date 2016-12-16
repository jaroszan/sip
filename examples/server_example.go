package main

import (
	"fmt"
	"github.com/jaroszan/sip"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Packet stores payload + destination address for the payload
type Packet struct {
	addr *net.UDPAddr
	data []byte
}

const udpPacketSize = 2048

func handleIncomingPacket(inbound chan Packet, outbound chan Packet, wg *sync.WaitGroup) {
	defer wg.Done()
	for packet := range inbound {
		log.Println("1")
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
	connection, _ := sip.Start("127.0.0.1:5160")

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

	// Prepare peer data
	remotePeerAddr, err := net.ResolveUDPAddr("udp", "localhost:5060")
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Millisecond * 300)
	go func() {
		for _ = range ticker.C {
			// Prepare INVITE
			newRequest := sip.MakeRequest("INVITE")
			outbound <- Packet{remotePeerAddr, []byte(newRequest)}
		}
	}()
	time.Sleep(time.Second * 1)
	ticker.Stop()
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
		b := make([]byte, udpPacketSize)
		n, addr, err := connection.ReadFromUDP(b)
		if err != nil {
			log.Println("Error on read: ", err)
			continue
		}
		inbound <- Packet{addr, b[:n]}
	}
}
