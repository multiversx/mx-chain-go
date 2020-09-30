package peerDiscovery

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

var durationMsgReceived = 2 * time.Second

// RunTest will test if all the peers receive a message
func RunTest(peers []p2p.Messenger, testIndex int, topic string) bool {
	fmt.Printf("Running test %v\n", testIndex)

	testMessage := "test " + strconv.Itoa(testIndex)
	messageProcessors := make([]*MessageProcesssor, len(peers))

	chanDone := make(chan struct{})
	chanMessageProcessor := make(chan struct{}, len(peers))

	//add a new message processor for each messenger
	for i, peer := range peers {
		mp := NewMessageProcessor(chanMessageProcessor, []byte(testMessage))

		messageProcessors[i] = mp
		err := peer.RegisterMessageProcessor(topic, "test", mp)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}

	var msgReceived int32 = 0

	go func() {

		for {
			<-chanMessageProcessor

			completelyRecv := true

			atomic.StoreInt32(&msgReceived, 0)

			//to be 100% all peers received the messages, iterate all message processors and check received flag
			for _, mp := range messageProcessors {
				if !mp.WasDataReceived() {
					completelyRecv = false
					continue
				}

				atomic.AddInt32(&msgReceived, 1)
			}

			if !completelyRecv {
				continue
			}

			//all messengers got the message
			chanDone <- struct{}{}
			return
		}
	}()

	//write the message on topic
	peers[0].Broadcast(topic, []byte(testMessage))

	select {
	case <-chanDone:
		return true
	case <-time.After(durationMsgReceived):
		fmt.Printf("timeout fetching all messages. Got %d from %d\n",
			atomic.LoadInt32(&msgReceived), len(peers))
		return false
	}
}
