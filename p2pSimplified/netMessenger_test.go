package components

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type interceptorStub struct {
	ProcessMessageCalled func(msg *pubsub.Message) error
}

// ProcessMessage -
func (stub *interceptorStub) ProcessMessage(msg *pubsub.Message) error {
	if stub.ProcessMessageCalled != nil {
		return stub.ProcessMessageCalled(msg)
	}

	return nil
}

func TestNewNetMessenger(t *testing.T) {
	t.Run("empty key bytes & empty initial peer list should not panic", func(t *testing.T) {
		args := ArgsNetMessenger{
			InitialPeerList: nil,
			PrivateKeyBytes: nil,
			ProtocolID:      "test",
		}

		netMes, err := NewNetMessenger(args)
		assert.NotNil(t, netMes)
		assert.Nil(t, err)

		_ = netMes.Close()
	})
	t.Run("provided key bytes should not panic", func(t *testing.T) {
		privateKeyBytes, err := GeneratePrivateKeyBytes()
		assert.Nil(t, err)

		fmt.Printf("generated private key bytes %x\n", privateKeyBytes)
		args := ArgsNetMessenger{
			InitialPeerList: nil,
			PrivateKeyBytes: privateKeyBytes,
			ProtocolID:      "test",
		}

		netMes, err := NewNetMessenger(args)
		assert.NotNil(t, netMes)
		assert.Nil(t, err)

		_ = netMes.Close()
	})
}

func TestNetMessenger_BootstrapConnectedPeers(t *testing.T) {
	nodes := make([]*netMessenger, 0)
	defer func() {
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("starting seednode...")
	argsSeeder := ArgsNetMessenger{
		InitialPeerList: nil,
		PrivateKeyBytes: nil,
		ProtocolID:      "test",
	}
	seeder, _ := NewNetMessenger(argsSeeder)
	nodes = append(nodes, seeder)
	seeder.Bootstrap()
	time.Sleep(time.Second)

	initialPeerList := []string{getLocalhostAddress(seeder)}
	argsNodes := ArgsNetMessenger{
		InitialPeerList: initialPeerList,
		PrivateKeyBytes: nil, // generate private key for each instance
		ProtocolID:      "test",
	}

	numNodes := 4
	fmt.Printf("starting %d nodes...\n", numNodes)
	for i := 0; i < numNodes; i++ {
		node, err := NewNetMessenger(argsNodes)
		require.Nil(t, err)

		nodes = append(nodes, node)
		node.Bootstrap()
	}

	time.Sleep(time.Second * 5)

	// test all nodes are connected
	for _, n1 := range nodes {
		for _, n2 := range nodes {
			// not testing self connections
			if n1.ID() == n2.ID() {
				continue
			}

			assert.Equal(t, network.Connected, n1.GetConnectedness(n2.ID()))
		}
	}
}

func getLocalhostAddress(netMes *netMessenger) string {
	addresses := netMes.Addresses()
	for _, addr := range addresses {
		if strings.Contains(addr, "127.0.0.1") {
			return addr
		}
	}

	return "not found"
}

func TestNetMessenger_BroadcastShouldwork(t *testing.T) {
	nodes := make([]*netMessenger, 0)
	defer func() {
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("starting seednode...")
	argsSeeder := ArgsNetMessenger{
		InitialPeerList: nil,
		PrivateKeyBytes: nil,
		ProtocolID:      "test",
	}
	seeder, _ := NewNetMessenger(argsSeeder)
	nodes = append(nodes, seeder)
	seeder.Bootstrap()
	time.Sleep(time.Second)

	initialPeerList := []string{getLocalhostAddress(seeder)}
	argsNodes := ArgsNetMessenger{
		InitialPeerList: initialPeerList,
		PrivateKeyBytes: nil, // generate private key for each instance
		ProtocolID:      "test",
	}

	numNodes := 4
	fmt.Printf("starting %d nodes...\n", numNodes)
	for i := 0; i < numNodes; i++ {
		node, err := NewNetMessenger(argsNodes)
		require.Nil(t, err)

		nodes = append(nodes, node)
		node.Bootstrap()
	}

	time.Sleep(time.Second * 5)

	mutMapDataRecieved := sync.RWMutex{}
	mapDataReceived := make(map[string]int)

	topic := "topic"
	message := []byte("message")

	// attach message processors
	for i, node := range nodes {
		if i == 0 {
			continue // except the seeder
		}

		processor := &interceptorStub{
			ProcessMessageCalled: func(msg *pubsub.Message) error {
				mutMapDataRecieved.Lock()
				mapDataReceived[string(msg.Data)]++
				mutMapDataRecieved.Unlock()

				return nil
			},
		}

		err := node.CreateTopic(topic)
		require.Nil(t, err)

		err = node.RegisterMessageProcessor(topic, processor)
		require.Nil(t, err)
	}

	// wait for topic broadcast
	time.Sleep(time.Second)

	nodes[1].Broadcast(topic, message)

	// wait for message broadcast
	time.Sleep(time.Second)

	mutMapDataRecieved.RLock()
	assert.Equal(t, len(nodes)-1, mapDataReceived[string(message)])
	mutMapDataRecieved.RUnlock()
}
