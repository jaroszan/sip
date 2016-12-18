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
		go func() {
			payload := string(packet.data)
			lines := strings.Split(payload, "\n")
			mType, mValue, err := sip.ParseFirstLine(lines[0])

			if err != nil {
				fmt.Println(err)
			}

			if mType == sip.REQUEST {
				sipHeaders := sip.ParseHeaders(lines[1:])
				if mValue == "INVITE" {
					outboundTrying := sip.PrepareResponse(sipHeaders, 100, "Trying")
					outbound180 := sip.PrepareResponse(sipHeaders, 180, "Ringing")
					outbound180 = sip.AddHeader(outbound180, "Contact", "sip:bob@localhost:5060")
					outboundOK := sip.PrepareResponse(sipHeaders, 200, "OK")
					outboundOK = sip.AddHeader(outboundOK, "Contact", "sip:alice@localhost:5060")
					outbound <- Packet{packet.addr, []byte(outboundTrying)}
					outbound <- Packet{packet.addr, []byte(outbound180)}
					outbound <- Packet{packet.addr, []byte(outboundOK)}
				} else if mValue == "BYE" {
					outboundOK := sip.PrepareResponse(sipHeaders, 200, "OK")
					outbound <- Packet{packet.addr, []byte(outboundOK)}
				} else {
					log.Println(mValue + " received")
				}
			} else if mType == sip.RESPONSE {
				if mValue == "200" {
					log.Println("200 OK received")
					sipHeaders := sip.ParseHeaders(lines[1:])
					if sipHeaders["cseq"] == "1 INVITE" {
						ackRequest := sip.MakeSubsequentRequest("ACK", "1", sipHeaders)
						outbound <- Packet{packet.addr, []byte(ackRequest)}
						byeRequest := sip.MakeSubsequentRequest("BYE", "2", sipHeaders)
						time.Sleep(time.Second * 2)
						outbound <- Packet{packet.addr, []byte(byeRequest)}
					}
				} else if mValue < "200" {
					log.Println("Provisional response received: " + mValue)
				} else {
					log.Println("Response received: " + mValue)
				}
			}
		}()
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

	ticker := time.NewTicker(time.Millisecond * 50)
	go func() {
		for _ = range ticker.C {
			// Prepare INVITE
			newRequest := sip.MakeRequest("INVITE", "1")
			outbound <- Packet{remotePeerAddr, []byte(newRequest)}
		}
	}()
	time.Sleep(time.Second * 90)
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
