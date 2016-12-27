package main

import (
	"github.com/jaroszan/sip"
	"log"
	"runtime"
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

func handleIncomingPacket(inbound chan sip.SipMessage, outbound chan sip.SipMessage) {
	for sipMessage := range inbound {
		sipMessage := sipMessage
		go func() {
			mType, mValue, err := sip.ParseFirstLine(sipMessage.FirstLine)
			if err != nil {
				log.Println(err)
				log.Println("Dropping request")
				runtime.Goexit()
			}

			if mType == sip.REQUEST {
				//PROCESS REQUESTS HERE
			} else if mType == sip.RESPONSE {
				mu.Lock()
				if _, ok := existingSessions[sipMessage.Headers["call-id"]]; !ok {
					existingSessions[sipMessage.Headers["call-id"]] = sessionData{0}
				}
				mu.Unlock()
				if mValue == "200" {
					if sipMessage.Headers["cseq"] == "1 INVITE" {
						mu.Lock()
						isOkReceived := existingSessions[sipMessage.Headers["call-id"]].ReceivedOK
						mu.Unlock()
						if isOkReceived == 0 {
							mu.Lock()
							existingSessions[sipMessage.Headers["call-id"]] = sessionData{1}
							mu.Unlock()
							ackRequest := sip.PrepareInDialogRequest("ACK", "1", sipMessage.Headers)
							outbound <- ackRequest
							byeRequest := sip.PrepareInDialogRequest("BYE", "2", sipMessage.Headers)
							time.Sleep(time.Second * 2)
							outbound <- byeRequest
						} else {
							log.Println("Retransmission received")
						}
					} else if sipMessage.Headers["cseq"] == "2 BYE" {
						mu.Lock()
						delete(existingSessions, sipMessage.Headers["call-id"])
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
	// Define local and remote peer
	localAddr := "localhost:5060"
	remoteAddr := "localhost:5160"
	// Define protocol to be used, either TCP or UDP is a valid choice
	transport := "UDP"
	// Initiate TCP connection to remote peer, inbound/outbound are channels are used
	// for receiving and sending messages respectively
	inbound, outbound := sip.StartSIP(localAddr, remoteAddr, transport)
	// Goroutine for processing incoming messages
	go handleIncomingPacket(inbound, outbound)

	ticker := time.NewTicker(time.Millisecond * 25)
	go func() {
		for _ = range ticker.C {
			// Prepare INVITE
			newRequest := sip.NewDialog("sip:bob@"+localAddr, "sip:alice@"+remoteAddr, transport)
			outbound <- newRequest
		}
	}()
	time.Sleep(time.Millisecond * 30)
	ticker.Stop()
	time.Sleep(time.Second * 5)
}
