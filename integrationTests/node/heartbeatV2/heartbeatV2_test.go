package heartbeatV2

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

func TestHeartbeatV2_AllPeersSendMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.NewTestHeartbeatNode(3, 0, interactingNodes)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 5)

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
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.NewTestHeartbeatNode(3, 0, interactingNodes)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 5)

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5
	checkMessages(t, nodes, maxMessageAgeAllowed)

	// Add new delayed node which requests messages
	delayedNode := integrationTests.NewTestHeartbeatNode(3, 0, interactingNodes+1)
	nodes = append(nodes, delayedNode)
	connectNodes(nodes, len(nodes))
	// Wait for messages to broadcast and requests to finish
	time.Sleep(time.Second * 5)

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}

	// Check sent messages again - now should have from all peers
	maxMessageAgeAllowed = time.Second * 5 // should not have messages from first Send
	checkMessages(t, nodes, maxMessageAgeAllowed)
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
			assert.True(t, paCache.Has(node.Messenger.ID().Bytes()))
			assert.True(t, hbCache.Has(node.Messenger.ID().Bytes()))

			// Also check message age
			value, _ := paCache.Get(node.Messenger.ID().Bytes())
			msg := value.(heartbeat.PeerAuthentication)

			marshaller := integrationTests.TestMarshaller
			payload := &heartbeat.Payload{}
			err := marshaller.Unmarshal(payload, msg.Payload)
			assert.Nil(t, err)

			currentTimestamp := time.Now().Unix()
			messageAge := time.Duration(currentTimestamp - payload.Timestamp)
			assert.True(t, messageAge < maxMessageAgeAllowed)
		}
	}
}
