package main

import (
	//"bufio"
	"fmt"
	"github.com/jaroszan/sip"
	"log"
	//"net"
	//"os"
	//"strconv"
	//"bytes"
	"strings"
	"sync"
	"time"
)

//Create mutex to protect existingSessions
var mu = &sync.RWMutex{}

type sessionData struct {
	ReceivedOK uint8
}

var existingSessions map[string]sessionData

func init() {
	existingSessions = make(map[string]sessionData)
}

func handleIncomingPacket(inbound chan []byte, outbound chan []byte) {
	//defer wg.Done()
	for message := range inbound {
		message := message
		go func() {
			var requestHandled bool
			requestHandled = false
			payload := string(message)
			lines := strings.Split(payload, "\n")
			mType, mValue, err := sip.ParseFirstLine(lines[0])

			if err != nil {
				fmt.Println(err)
			}
			sipHeaders := sip.ParseHeaders(lines[1:])

			if mType == sip.REQUEST {
				sipHeaders := sip.ParseHeaders(lines[1:])
				if mValue == "INVITE" {
					outboundTrying := sip.PrepareResponse(sipHeaders, 100, "Trying")
					outbound180 := sip.PrepareResponse(sipHeaders, 180, "Ringing")
					outbound180 = sip.AddHeader(outbound180, "Contact", "sip:bob@localhost:5060")
					outboundOK := sip.PrepareResponse(sipHeaders, 200, "OK")
					outboundOK = sip.AddHeader(outboundOK, "Contact", "sip:alice@localhost:5060")
					outbound <- []byte(outboundTrying)
					outbound <- []byte(outbound180)
					outbound <- []byte(outboundOK)
				} else if mValue == "BYE" {
					outboundOK := sip.PrepareResponse(sipHeaders, 200, "OK")
					outbound <- []byte(outboundOK)
				} else {
					log.Println(mValue + " received")
				}
			} else if mType == sip.RESPONSE {
				mu.Lock()
				if _, ok := existingSessions[sipHeaders["call-id"]]; !ok {
					existingSessions[sipHeaders["call-id"]] = sessionData{0}
				}
				mu.Unlock()
				if mValue == "200" {
					if sipHeaders["cseq"] == "1 INVITE" {
						mu.Lock()
						isOkReceived := existingSessions[sipHeaders["call-id"]].ReceivedOK
						mu.Unlock()
						if requestHandled == false && isOkReceived == 0 {
							mu.Lock()
							existingSessions[sipHeaders["call-id"]] = sessionData{1}
							mu.Unlock()
							requestHandled = true
							ackRequest := sip.PrepareInDialogRequest("ACK", "1", sipHeaders)
							outbound <- []byte(ackRequest)
							byeRequest := sip.PrepareInDialogRequest("BYE", "2", sipHeaders)
							time.Sleep(time.Second * 2)
							outbound <- []byte(byeRequest)
						} else {
							log.Println("Retransmission received")
						}
					} else if sipHeaders["cseq"] == "2 BYE" {
						mu.Lock()
						delete(existingSessions, sipHeaders["call-id"])
						mu.Unlock()
					}
				} else if mValue < "200" {
					//log.Println("Provisional response received: " + mValue)
				} else {
					log.Println("Response received: " + mValue)
				}
			}
		}()
	}
}

func main() {
	// Initiate TCP connection to remote peer, inbound/outbound are channels are used
	// for receiving and sending messages respectively
	inbound, outbound, conn := sip.StartTCPClient("127.0.0.1:5160", "127.0.0.1:5060")
	defer conn.Close()
	// Goroutine for processing incoming messages
	go handleIncomingPacket(inbound, outbound)

	ticker := time.NewTicker(time.Millisecond * 40)
	go func() {
		for _ = range ticker.C {
			// Prepare INVITE
			newRequest := sip.NewDialog("sip:bob@localhost:5160", "sip:alice@localhost:5060", "TCP")
			outbound <- []byte(newRequest)
		}
	}()
	time.Sleep(time.Second * 300)
	ticker.Stop()
	//conn.Close()
	time.Sleep(time.Second * 5)
}
