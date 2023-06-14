package heartbeatV2

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/integrationTests"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("inteagrationtests/node/heartbeatv2")

func TestHeartbeatV2_AllPeersSendMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 60)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5
	checkMessages(t, nodes, maxMessageAgeAllowed)
}

func TestHeartbeatV2_PeerJoiningLate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 60)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5
	checkMessages(t, nodes, maxMessageAgeAllowed)

	// Add new delayed node which requests messages
	delayedNode := integrationTests.NewTestHeartbeatNode(t, 3, 0, 0, p2pConfig, 60)
	nodes = append(nodes, delayedNode)
	connectNodes(nodes, len(nodes))
	// Wait for messages to broadcast and requests to finish
	time.Sleep(time.Second * 15)

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}

	// Check sent messages again - now should have from all peers
	maxMessageAgeAllowed = time.Second * 5 // should not have messages from first Send
	checkMessages(t, nodes, maxMessageAgeAllowed)
}

func TestHeartbeatV2_PeerAuthenticationMessageExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 20)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	log.Info("wait for messages to broadcast...")
	time.Sleep(time.Second * 15)

	log.Info("closing the last node")
	lastNode := nodes[interactingNodes-1]
	lastNode.Close()

	log.Info("waiting the messages from the last node expire")
	time.Sleep(time.Second * 30)

	log.Info("first node clears its peer authentication messages data pool after all other nodes stop sending peer authentication messages")
	for i := 1; i < len(nodes)-1; i++ {
		_ = nodes[i].Sender.Close()
	}
	time.Sleep(time.Second * 2)
	nodes[0].DataPool.PeerAuthentications().Clear()

	log.Info("first node requests the peer authentication messages from other peer(s)")

	requestHashes := make([][]byte, 0)
	for i := 1; i < len(nodes); i++ {
		pkBytes, err := nodes[i].NodeKeys.MainKey.Pk.ToByteArray()
		assert.Nil(t, err)

		requestHashes = append(requestHashes, pkBytes)
	}

	nodes[0].RequestHandler.RequestPeerAuthenticationsByHashes(nodes[0].ShardCoordinator.SelfId(), requestHashes)

	time.Sleep(time.Second * 5)

	// first node should have not received the requested message because it is expired
	lastPkBytes := requestHashes[len(requestHashes)-1]
	assert.False(t, nodes[0].RequestedItemsHandler.Has(string(lastPkBytes)))
	assert.Equal(t, interactingNodes-2, nodes[0].DataPool.PeerAuthentications().Len())
}

func connectNodes(nodes []*integrationTests.TestHeartbeatNode, interactingNodes int) {
	for i := 0; i < interactingNodes-1; i++ {
		for j := i + 1; j < interactingNodes; j++ {
			src := nodes[i]
			dst := nodes[j]
			_ = src.ConnectTo(dst)
		}
	}
}

func checkMessages(t *testing.T, nodes []*integrationTests.TestHeartbeatNode, maxMessageAgeAllowed time.Duration) {
	numOfNodes := len(nodes)
	for i := 0; i < numOfNodes; i++ {
		paCache := nodes[i].DataPool.PeerAuthentications()
		hbCache := nodes[i].DataPool.Heartbeats()

		assert.Equal(t, numOfNodes, paCache.Len())
		assert.Equal(t, numOfNodes, hbCache.Len())

		// Check this node received messages from all peers
		for _, node := range nodes {
			pkBytes, err := node.NodeKeys.MainKey.Pk.ToByteArray()
			assert.Nil(t, err)

			assert.True(t, paCache.Has(pkBytes))
			assert.True(t, hbCache.Has(node.MainMessenger.ID().Bytes()))

			// Also check message age
			value, found := paCache.Get(pkBytes)
			require.True(t, found)
			msg := value.(*heartbeat.PeerAuthentication)

			marshaller := integrationTests.TestMarshaller
			payload := &heartbeat.Payload{}
			err = marshaller.Unmarshal(payload, msg.Payload)
			assert.Nil(t, err)

			currentTimestamp := time.Now().Unix()
			messageAge := time.Duration(currentTimestamp - payload.Timestamp)
			assert.True(t, messageAge < maxMessageAgeAllowed)
		}
	}
}
