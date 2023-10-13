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

	t.Run("single initiator", testMessagePropagation(400, 133, 6, 1, 1000))
	t.Run("multiple initiators", testMessagePropagation(400, 133, 6, 5, 1000))
}

func testMessagePropagation(numNodes int, numOfMaliciousPeers int, numOfBroadcasts int, numInitiators int, numWorkers int) func(t *testing.T) {
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
			nodes[i], err = NewTestOnlySyncedMessenger(network, seqNoGenerator, isMalicious)
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

		workerPoolInstance.StopWait()

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

			if isMalicious {
				assert.Equal(t, 0, msgCnt.sent, fmt.Sprintf("%s @idx %d malicious should not send any message", pid.Pretty(), i))
				continue
			}

			assert.Equal(t, 1, msgCnt.sent, fmt.Sprintf("%s @idx %d didn't send any message", pid.Pretty(), i))
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
	}
}
