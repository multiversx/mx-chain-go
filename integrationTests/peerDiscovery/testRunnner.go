package peerDiscovery

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

var durationMsgRecieved = time.Duration(time.Second * 2)

type TestRunner struct {
}

func (tr *TestRunner) CreateMessenger(ctx context.Context,
	port int,
	peerDiscoverer p2p.PeerDiscoverer) p2p.Messenger {

	r := rand.New(rand.NewSource(int64(port)))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessenger(
		ctx,
		port,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		peerDiscoverer)

	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func (tr *TestRunner) RunTest(peers []p2p.Messenger, testIndex int, topic string) bool {
	fmt.Printf("Running test %v\n", testIndex)

	testMessage := "test " + strconv.Itoa(testIndex)
	messageProcessors := make([]*MessageProcesssor, len(peers))

	chanDone := make(chan struct{})
	chanMessageProcessor := make(chan struct{}, len(peers))

	//add a new message processor for each messenger
	for i, peer := range peers {
		if peer.HasTopicValidator(topic) {
			_ = peer.UnregisterMessageProcessor(topic)
		}

		mp := NewMessageProcessor(chanMessageProcessor, []byte(testMessage))

		messageProcessors[i] = mp
		err := peer.RegisterMessageProcessor(topic, mp)
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
	case <-time.After(durationMsgRecieved):
		fmt.Printf("timeout fetching all messages. Got %d from %d\n",
			atomic.LoadInt32(&msgReceived), len(peers))
		return false
	}
}
