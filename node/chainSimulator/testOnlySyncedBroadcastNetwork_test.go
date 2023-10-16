package chainSimulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	chainP2P "github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic                 = "topic"
	message               = "message"
	p2pBootstrapStepDelay = time.Minute * 5
)

func TestTestOnlySyncedBroadcastNetwork_EquivalentMessagesRealMessengers(t *testing.T) {
	t.Skip("testing only")

	t.Run("single initiator, equivalent messages filter", testMessagePropagationRealMessenger(800, 266, 1))
	t.Run("multiple initiators, equivalent messages filter", testMessagePropagationRealMessenger(800, 266, 5))
}

func TestTestOnlySyncedBroadcastNetwork_EquivalentMessages(t *testing.T) {
	t.Skip("testing only")

	t.Run("single initiator, no equivalent messages filter", testMessagePropagation(400, 133, 6, 1, 1000, false))
	t.Run("multiple initiators, no equivalent messages filter", testMessagePropagation(400, 133, 6, 5, 1000, false))

	t.Run("single initiator, equivalent messages filter", testMessagePropagation(400, 133, 6, 1, 1000, true))
	t.Run("multiple initiators, equivalent messages filter", testMessagePropagation(400, 133, 6, 5, 1000, true))
}

func testMessagePropagationRealMessenger(numNodes int, numOfMaliciousPeers int, numInitiators int) func(t *testing.T) {
	return func(t *testing.T) {

		advertiser := integrationTests.CreateMessengerWithKadDht("")
		_ = advertiser.Bootstrap()
		seedAddress := integrationTests.GetConnectableAddress(advertiser)

		cfg := config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0",
				Transports: config.P2PTransportConfig{
					TCP: p2pConfig.TCPProtocolConfig{
						ListenAddress: p2p.LocalHostListenAddrWithIp4AndTcp,
					},
				},
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				Type:                             "optimized",
				RefreshIntervalInSec:             2,
				ProtocolID:                       "/erd/kad/1.0.0",
				BucketSize:                       100,
				RoutingTableRefreshIntervalInSec: 100,
				InitialPeerList:                  []string{seedAddress},
			},
			Sharding: config.ShardingConfig{
				TargetPeerCount: 40,
				Type:            p2p.OneListSharder,
			},
		}

		nodes := make([]chainP2P.Messenger, numNodes)
		for i := 0; i < numNodes; i++ {
			nodes[i] = integrationTests.CreateMessengerFromConfig(cfg)
		}

		startingIndexOfMaliciousPeers := numNodes - numOfMaliciousPeers
		processors := createTestInterceptorForEachNode(nodes, startingIndexOfMaliciousPeers)

		time.Sleep(time.Second * 2)

		startNodes(nodes)

		fmt.Println("Delaying for node bootstrap and topic announcement...")
		time.Sleep(p2pBootstrapStepDelay)

		for i := 0; i < numInitiators; i++ {
			nodes[i].Broadcast(topic, []byte(message))
			time.Sleep(time.Millisecond * 100)
		}

		time.Sleep(time.Second * 10)

		fmt.Println("Closing the nodes...")
		stopNodes(advertiser, nodes)

		cntReceivedMessages := 0
		maxMessagesReceived := 0
		cntMissedNodes := 0
		for i := 0; i < len(processors); i++ {
			seenMessages := processors[i].GetSeenMessages()
			msgCnt, messageReceived := seenMessages[message]
			if !messageReceived {
				cntMissedNodes++
				continue
			}

			isMalicious := processors[i].IsMalicious()
			pid := nodes[i].ID()
			println(fmt.Sprintf("%s equivalent messages stats: %d sent, %d received, isMalicious: %t", pid.Pretty(), msgCnt.sent, msgCnt.received, isMalicious))

			cntReceivedMessages += msgCnt.received
			if msgCnt.received > maxMessagesReceived {
				maxMessagesReceived = msgCnt.received
			}

			expectedSent := 1
			if isMalicious {
				expectedSent = 0
			}

			assert.Equal(t, expectedSent, msgCnt.sent, fmt.Sprintf("%s @idx %d didn't send any message, isMalicious %t", pid.Pretty(), i, isMalicious))
		}
		assert.Equal(t, 0, cntMissedNodes, "all nodes should have received the message")

		println(fmt.Sprintf("\nTest info: %d nodes, %d malicious, %d initiators",
			numNodes, numOfMaliciousPeers, numInitiators,
		))

		println(fmt.Sprintf("Results equivalent messages:\n"+
			"\t- max messages received by a peer %d\n"+
			"\t- average messages received by a peer %f",
			maxMessagesReceived, float64(cntReceivedMessages)/float64(numNodes)))
	}
}

func startNodes(nodes []p2p.Messenger) {
	for _, node := range nodes {
		_ = node.Bootstrap()
	}
}

func stopNodes(advertiser p2p.Messenger, nodes []p2p.Messenger) {
	_ = advertiser.Close()
	for _, node := range nodes {
		_ = node.Close()
	}
}

func createTestInterceptorForEachNode(nodes []p2p.Messenger, startingIndexOfMaliciousNodes int) []*testOnlyMessageProcessor {
	processors := make([]*testOnlyMessageProcessor, len(nodes))
	for i := 0; i < len(nodes); i++ {
		_ = nodes[i].CreateTopic(topic, true)
		isMalicious := i >= startingIndexOfMaliciousNodes
		processor, _ := NewTestOnlyMessageProcessor(nodes[i], isMalicious)
		processors[i] = processor
		_ = nodes[i].RegisterMessageProcessor(topic, "", processor)
	}

	return processors
}

func testMessagePropagation(numNodes int, numOfMaliciousPeers int, numOfBroadcasts int, numInitiators int, numWorkers int, equivalentMessagesFilterOn bool) func(t *testing.T) {
	return func(t *testing.T) {
		// workerPoolInstance keeps the go routines created by broadcasts in control
		workerPoolInstance := workerpool.New(numWorkers)
		// chanFinish will be called when all nodes received the message at least once
		chanFinish := make(chan bool, 1)
		network, err := NewTestOnlySyncedBroadcastNetwork(numOfBroadcasts, workerPoolInstance, chanFinish)
		require.Nil(t, err)

		seqNoGenerator := NewSequenceGenerator()
		nodes := make([]*testOnlySyncedMessenger, numNodes)
		startingIndexOfMaliciousPeers := numNodes - numOfMaliciousPeers
		for i := 0; i < numNodes; i++ {
			isMalicious := i >= startingIndexOfMaliciousPeers
			nodes[i], err = NewTestOnlySyncedMessenger(network, seqNoGenerator, isMalicious, equivalentMessagesFilterOn)
			require.Nil(t, err)

			_ = nodes[i].CreateTopic(topic, true)
			_ = nodes[i].RegisterMessageProcessor(topic, "", nodes[i])
		}

		for i := 0; i < numInitiators; i++ {
			nodes[i].Broadcast(topic, []byte(message))
		}

		done := false
		start := time.Now()
		for !done {
			select {
			case <-chanFinish:
				done = true
			case <-time.After(time.Second * 10):
				assert.Fail(t, "timeout")
				done = true
			}
		}
		duration := time.Since(start)

		uniqueMessagesTotal := make(map[string]int)

		cntReceivedMessages := 0
		maxMessagesReceived := 0
		cntMissedNodes := 0
		for i := 0; i < numNodes; i++ {
			seenMessages := nodes[i].getSeenMessages()
			msgCnt, messageReceived := seenMessages[message]
			if !messageReceived {
				cntMissedNodes++
				continue
			}

			pid := nodes[i].ID()
			isMalicious := i >= startingIndexOfMaliciousPeers
			println(fmt.Sprintf("%s equivalent messages stats: %d sent, %d received, isMalicious: %t", pid.Pretty(), msgCnt.sent, msgCnt.received, isMalicious))

			cntReceivedMessages += msgCnt.received
			if msgCnt.received > maxMessagesReceived {
				maxMessagesReceived = msgCnt.received
			}

			uniqueMessages := nodes[i].getUniqueMessages()
			println(fmt.Sprintf("%s unique messages stats:", pid.Pretty()))
			for key, cnt := range uniqueMessages {
				uniqueMessagesTotal[key] += cnt
				println(fmt.Sprintf("\t- key: %s, %d times received", key, cnt))
			}

			// skip checks for filter off, nodes may send 1->numInitiators messages
			if !equivalentMessagesFilterOn {
				continue
			}

			expectedSent := 1
			if isMalicious {
				expectedSent = 0
			}

			assert.Equal(t, expectedSent, msgCnt.sent, fmt.Sprintf("%s @idx %d didn't send any message, isMalicious %t", pid.Pretty(), i, isMalicious))
		}

		assert.Equal(t, 0, cntMissedNodes, "all nodes should have received the message")

		println(fmt.Sprintf("\nTest info: %d nodes, %d malicious, %d broadcasts, %d initiators",
			numNodes, numOfMaliciousPeers, numOfBroadcasts, numInitiators,
		))

		println(fmt.Sprintf("Results equivalent messages:\n"+
			"\t- message reached all nodes after %s\n"+
			"\t- max messages received by a peer %d\n"+
			"\t- average messages received by a peer %f",
			duration, maxMessagesReceived, float64(cntReceivedMessages)/float64(numNodes)))

		println("Results unique messages:")
		for key, total := range uniqueMessagesTotal {
			println(fmt.Sprintf("\t- %s was received in total of %d times, with an average of %f times per node", key, total, float64(total)/float64(numNodes)))
		}
		println()

		workerPoolInstance.Stop()
	}
}
