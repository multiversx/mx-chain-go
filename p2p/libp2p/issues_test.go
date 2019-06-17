package libp2p_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
)

func createMessenger(port int) p2p.Messenger {
	r := rand.New(rand.NewSource(int64(port)))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

// TestIssueEN898_StreamResetError emphasizes what happens if direct sender writes to a stream that has been reset
// Testing is done by writing a large buffer that will cause the recipient to reset its inbound stream
// Sender will then be notified that the stream writing did not succeed but it will only log the error
// Next message that the sender tries to send will cause a new error to be logged and no data to be sent
// The fix consists in the full stream closing when an error occurs during writing.
func TestIssueEN898_StreamResetError(t *testing.T) {
	mes1 := createMessenger(23100)
	mes2 := createMessenger(23101)

	defer func() {
		mes1.Close()
		mes2.Close()
	}()

	mes1.ConnectToPeer(getConnectableAddress(mes2))

	topic := "test topic"

	size4MB := 1 << 22
	size4kB := 1 << 12

	//a 4MB slice containing character A
	largePacket := bytes.Repeat([]byte{65}, size4MB)
	//a 4kB slice containing character B
	smallPacket := bytes.Repeat([]byte{66}, size4kB)

	largePacketReceived := &atomic.Value{}
	largePacketReceived.Store(false)

	smallPacketReceived := &atomic.Value{}
	smallPacketReceived.Store(false)

	mes2.CreateTopic(topic, false)
	mes2.RegisterMessageProcessor(topic, &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P) error {
			if bytes.Equal(message.Data(), largePacket) {
				largePacketReceived.Store(true)
			}

			if bytes.Equal(message.Data(), smallPacket) {
				smallPacketReceived.Store(true)
			}

			return nil
		},
	})

	fmt.Println("sending the large packet...")
	mes1.SendToConnectedPeer(topic, largePacket, mes2.ID())

	time.Sleep(time.Second)

	fmt.Println("sending the small packet...")
	mes1.SendToConnectedPeer(topic, smallPacket, mes2.ID())

	time.Sleep(time.Second)

	assert.False(t, largePacketReceived.Load().(bool))
	assert.True(t, smallPacketReceived.Load().(bool))
}
