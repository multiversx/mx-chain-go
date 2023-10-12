package chainSimulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic   = "topic"
	message = "message"
)

func TestTestOnlySyncedBroadcastNetwork_EquivalentMessages(t *testing.T) {
	t.Skip("testing only")

	t.Run("single initiator", testMessagePropagation(1000, 10, 1, 1000))
	t.Run("multiple initiators", testMessagePropagation(1000, 10, 5, 1000))
}

func testMessagePropagation(numNodes int, numPeersPerNode int, numInitiators int, numWorkers int) func(t *testing.T) {
	return func(t *testing.T) {
		// workerPoolInstance keeps the go routines created by broadcasts in control
		workerPoolInstance := workerpool.New(numWorkers)
		// chanFinish will be called when all nodes received the message at least once
		chanFinish := make(chan bool, 1)
		network, err := NewTestOnlySyncedBroadcastNetwork(numPeersPerNode, workerPoolInstance, chanFinish)
		require.Nil(t, err)

		seqNoGenerator := NewSequenceGenerator()
		nodes := make([]*testOnlySyncedMessenger, numNodes)
		for i := 0; i < numNodes; i++ {
			nodes[i], err = NewTestOnlySyncedMessenger(network, seqNoGenerator)
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
			case <-time.After(time.Second * 30):
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
			println(fmt.Sprintf("%s equivalent messages stats: %d sent, %d received", pid.Pretty(), msgCnt.sent, msgCnt.received))

			cntReceivedMessages += msgCnt.received
			if msgCnt.received > maxMessagesReceived {
				maxMessagesReceived = msgCnt.received
			}

			uniqueMessages := nodes[i].getUniqueMessages()
			println(fmt.Sprintf("%s unique messages stats:", pid.Pretty()))
			for key, cnt := range uniqueMessages {
				uniqueMessagesTotal[key] += cnt
				println(fmt.Sprintf("      key: %s, %d times received", key, cnt))
			}

			require.Equal(t, 1, msgCnt.sent, fmt.Sprintf("%s @idx %d didn't send any message", pid.Pretty(), i))
		}

		require.Equal(t, 0, cntMissedNodes, "all nodes should have received the message")

		println(fmt.Sprintf("Results: %d nodes, %d peers, %d initiators\n"+
			"message reached all nodes after %s\n"+
			"max messages received by a peer %d\n"+
			"average messages received by a peer %f\n",
			numNodes, numPeersPerNode, numInitiators, duration, maxMessagesReceived, float64(cntReceivedMessages)/float64(numNodes)))

		println(fmt.Sprintf("Results unique messages:"))
		for key, total := range uniqueMessagesTotal {
			println(fmt.Sprintf("%s was received in total of %d times, with an average of %f times per node", key, total, float64(total)/float64(numNodes)))
		}

		workerPoolInstance.Stop()
	}
}
