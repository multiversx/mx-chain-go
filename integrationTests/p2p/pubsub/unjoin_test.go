package peerDisconnecting

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

const durationBootstrapping = time.Second * 2
const durationTraverseNetwork = time.Second * 2
const durationUnjoin = time.Second * 2

func TestPubsubUnjoinShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	peers, _ := integrationTests.CreateFixedNetworkOf8Peers()
	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	topic := "test_topic"
	processors := make([]*messageProcessor, 0, len(peers))
	for idx, p := range peers {
		_ = p.CreateTopic(topic, true)
		processors = append(processors, newMessageProcessor())
		_ = p.RegisterMessageProcessor(topic, "test", processors[idx])
	}

	fmt.Println("bootstrapping nodes")
	time.Sleep(durationBootstrapping)

	//a message should traverse the network
	fmt.Println("sending the message that should traverse the whole network")
	sender := peers[4]
	sender.Broadcast(topic, []byte("message 1"))

	time.Sleep(durationTraverseNetwork)

	for _, mp := range processors {
		assert.Equal(t, 1, len(mp.AllMessages()))
	}

	blockedIdxs := []int{3, 6, 2, 5}
	//node 3 unjoins the topic, which should prevent the propagation of the messages on peers 3, 6, 2 and 5
	err := peers[3].UnregisterAllMessageProcessors()
	assert.Nil(t, err)

	err = peers[3].UnjoinAllTopics()
	assert.Nil(t, err)

	time.Sleep(durationUnjoin)

	fmt.Println("sending the message that should traverse half the network")
	sender.Broadcast(topic, []byte("message 2"))

	time.Sleep(durationTraverseNetwork)

	for idx, mp := range processors {
		if integrationTests.IsIntInSlice(idx, blockedIdxs) {
			assert.Equal(t, 1, len(mp.AllMessages()))
			continue
		}

		assert.Equal(t, 2, len(mp.AllMessages()))
	}
}
