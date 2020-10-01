package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
)

func createMockNetworkArgs() libp2p.ArgsNetworkMessenger {
	return libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
	}
}

func main() {
	net := mocknet.New(context.Background())

	mes1, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), net)
	mes2, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), net)

	_ = net.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	_ = mes1.CreateTopic("test1", true)
	_ = mes1.CreateTopic("test2", true)
	_ = mes1.CreateTopic("test3", true)

	_ = mes2.CreateTopic("test1", true)
	_ = mes2.CreateTopic("test2", true)
	_ = mes2.CreateTopic("test3", true)

	bytesReceived1 := int64(0)
	bytesReceived2 := int64(0)
	bytesReceived3 := int64(0)

	_ = mes1.RegisterMessageProcessor("test1", "identifier",
		&mock.MessageProcessorStub{
			ProcessMessageCalled: func(message p2p.MessageP2P, _ core.PeerID) error {
				atomic.AddInt64(&bytesReceived1, int64(len(message.Data())))

				return nil
			},
		})

	_ = mes1.RegisterMessageProcessor("test2", "identifier", &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, _ core.PeerID) error {
			atomic.AddInt64(&bytesReceived2, int64(len(message.Data())))

			return nil
		},
	})

	_ = mes1.RegisterMessageProcessor("test3", "identifier", &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, _ core.PeerID) error {
			atomic.AddInt64(&bytesReceived3, int64(len(message.Data())))

			return nil
		},
	})

	time.Sleep(time.Second)

	timeStart := time.Now()
	bytesSent := int64(0)

	durTest := time.Second * 5

	fmt.Printf("Testing for %s...\n", durTest.String())

	for time.Now().UnixNano() < timeStart.Add(durTest).UnixNano() {
		buffSize := 5000
		buff := make([]byte, buffSize)
		bytesSent += int64(buffSize)

		mes2.Broadcast("test1", buff)
		mes2.Broadcast("test2", buff)
		//topic test3 receives more requests to send
		mes2.Broadcast("test3", buff)
		mes2.Broadcast("test3", buff)
	}

	fmt.Printf("Sent: %s -> %s\nReceived pipe 1 %s -> %s\nReceived pipe 2 %s -> %s\nReceived pipe 3 %s -> %s\n",
		bytesPretty(float64(bytesSent)), bytesPerSecPretty(bytesSent, durTest),
		bytesPretty(float64(bytesReceived1)), bytesPerSecPretty(bytesReceived1, durTest),
		bytesPretty(float64(bytesReceived2)), bytesPerSecPretty(bytesReceived2, durTest),
		bytesPretty(float64(bytesReceived3)), bytesPerSecPretty(bytesReceived3, durTest))
}

func bytesPretty(bytes float64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%.0f bytes", bytes)
	}

	if bytes < 1048576 {
		return fmt.Sprintf("%.2f kB", bytes/1024.0)
	}

	return fmt.Sprintf("%.2f MB", bytes/1048576.0)
}

func bytesPerSecPretty(bytes int64, dur time.Duration) string {
	return bytesPretty(float64(bytes)/dur.Seconds()) + "/s"
}
