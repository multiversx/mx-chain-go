package peerDisconnecting

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = 2 * time.Second
var durationTest = 30 * time.Second

type messageProcessorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

func (mps *messageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return mps.ProcessReceivedMessageCalled(message)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *messageProcessorStub) IsInterfaceNil() bool {
	if mps == nil {
		return true
	}
	return false
}

func TestPeerReceivesTheSameMessageMultipleTimesShouldNotHappen(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20

	//Step 1. Create advertiser
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")

	//Step 2. Create numOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		node := integrationTests.CreateMessengerWithKadDht(context.Background(),
			integrationTests.GetConnectableAddress(advertiser))
		peers[i] = node
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < numOfPeers; i++ {
			if peers[i] != nil {
				_ = peers[i].Close()
			}
		}

		if advertiser != nil {
			_ = advertiser.Close()
		}
	}()

	chanStop := make(chan struct{})

	//Step 3. Register pubsub validators
	mutMapMessages := sync.Mutex{}
	mapMessages := make(map[int]map[string]struct{})
	testTopic := "test"

	for i := 0; i < numOfPeers; i++ {
		idx := i
		mapMessages[idx] = make(map[string]struct{})
		err := peers[idx].CreateTopic(testTopic, true)
		if err != nil {
			fmt.Println("CreateTopic failed:", err.Error())
			continue
		}

		err = peers[idx].RegisterMessageProcessor(testTopic, &messageProcessorStub{
			ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
				time.Sleep(time.Second)

				mutMapMessages.Lock()
				defer mutMapMessages.Unlock()

				msgId := "peer: " + message.Peer().Pretty() + " - seqNo: 0x" + hex.EncodeToString(message.SeqNo())
				_, ok := mapMessages[idx][msgId]
				if ok {
					assert.Fail(t, "message %s received twice", msgId)
					chanStop <- struct{}{}
				}

				mapMessages[idx][msgId] = struct{}{}
				return nil
			},
		})
		if err != nil {
			fmt.Println("RegisterMessageProcessor:", err.Error())
		}
	}

	//Step 4. Call bootstrap on all peers
	err := advertiser.Bootstrap()
	if err != nil {
		fmt.Println("Bootstrap failed:", err.Error())
	}
	for _, p := range peers {
		err = p.Bootstrap()
		if err != nil {
			fmt.Println(fmt.Sprintf("Bootstrap() for peer id %s failed:%s", p.ID(), err.Error()))
		}
	}
	integrationTests.WaitForBootstrapAndShowConnected(peers, durationBootstrapingTime)

	//Step 5. Continuously send messages from one peer
	for timeStart := time.Now(); timeStart.Add(durationTest).Unix() > time.Now().Unix(); {
		peers[0].Broadcast(testTopic, []byte("test buff"))
		select {
		case <-chanStop:
			return
		default:
		}
		time.Sleep(time.Millisecond)
	}
}
