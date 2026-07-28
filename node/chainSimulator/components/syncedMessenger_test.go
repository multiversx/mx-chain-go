package components

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	p2pMessage "github.com/multiversx/mx-chain-communication-go/p2p/message"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
)

func TestNewSyncedMessenger(t *testing.T) {
	t.Parallel()

	t.Run("nil network should error", func(t *testing.T) {
		t.Parallel()

		messenger, err := NewSyncedMessenger(nil)
		assert.Nil(t, messenger)
		assert.Equal(t, errNilNetwork, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		messenger, err := NewSyncedMessenger(NewSyncedBroadcastNetwork())
		assert.NotNil(t, messenger)
		assert.Nil(t, err)
	})
}

func TestSyncedMessenger_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var messenger *syncedMessenger
	assert.True(t, messenger.IsInterfaceNil())

	messenger, _ = NewSyncedMessenger(NewSyncedBroadcastNetwork())
	assert.False(t, messenger.IsInterfaceNil())
}

func TestSyncedMessenger_DisabledMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

	assert.Nil(t, messenger.Close())
	assert.Zero(t, messenger.Port())
	assert.Nil(t, messenger.SetPeerDenialEvaluator(nil))
	assert.Nil(t, messenger.SetThresholdMinConnectedPeers(0))
	assert.Zero(t, messenger.ThresholdMinConnectedPeers())
	assert.True(t, messenger.IsConnectedToTheNetwork())
	assert.Nil(t, messenger.SetPeerShardResolver(nil))
	assert.Nil(t, messenger.ConnectToPeer(""))
	assert.Nil(t, messenger.Bootstrap())
	msgID, err := messenger.ProcessReceivedMessage(nil, "", nil)
	assert.Nil(t, err)
	assert.Nil(t, msgID)

	messenger.WaitForConnections(0, 0)

	buff, err := messenger.SignUsingPrivateKey(nil, nil)
	assert.Empty(t, buff)
	assert.Nil(t, err)
}

func TestSyncedMessenger_RegisterMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil message processor should error", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		err := messenger.RegisterMessageProcessor("", "", nil)
		assert.ErrorIs(t, err, errNilMessageProcessor)
	})
	t.Run("processor exists, should error", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		err := messenger.CreateTopic("t", false)
		assert.Nil(t, err)

		processor1 := &p2pmocks.MessageProcessorStub{}
		err = messenger.RegisterMessageProcessor("t", "i", processor1)
		assert.Nil(t, err)

		processor2 := &p2pmocks.MessageProcessorStub{}
		err = messenger.RegisterMessageProcessor("t", "i", processor2)
		assert.ErrorIs(t, err, errTopicHasProcessor)

		messenger.mutOperation.RLock()
		defer messenger.mutOperation.RUnlock()

		assert.True(t, messenger.topics["t"]["i"] == processor1) // pointer testing
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		err := messenger.CreateTopic("t", false)
		assert.Nil(t, err)

		processor := &p2pmocks.MessageProcessorStub{}
		err = messenger.RegisterMessageProcessor("t", "i", processor)
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		defer messenger.mutOperation.RUnlock()

		assert.True(t, messenger.topics["t"]["i"] == processor) // pointer testing
	})
}

func TestSyncedMessenger_UnregisterAllMessageProcessors(t *testing.T) {
	t.Parallel()

	t.Run("no topics should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())
		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics)
		messenger.mutOperation.RUnlock()

		err := messenger.UnregisterAllMessageProcessors()
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics)
		messenger.mutOperation.RUnlock()
	})
	t.Run("one topic but no processor should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		topic := "topic"
		_ = messenger.CreateTopic(topic, true)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics[topic])
		messenger.mutOperation.RUnlock()

		err := messenger.UnregisterAllMessageProcessors()
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics[topic])
		messenger.mutOperation.RUnlock()
	})
	t.Run("one topic with processor should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		topic := "topic"
		identifier := "identifier"
		_ = messenger.CreateTopic(topic, true)
		_ = messenger.RegisterMessageProcessor(topic, identifier, &p2pmocks.MessageProcessorStub{})

		messenger.mutOperation.RLock()
		assert.NotNil(t, messenger.topics[topic][identifier])
		messenger.mutOperation.RUnlock()

		err := messenger.UnregisterAllMessageProcessors()
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics[topic])
		messenger.mutOperation.RUnlock()
	})
}

func TestSyncedMessenger_UnregisterMessageProcessor(t *testing.T) {
	t.Parallel()

	t.Run("topic not found should error", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		topic := "topic"
		identifier := "identifier"
		err := messenger.UnregisterMessageProcessor(topic, identifier)
		assert.ErrorIs(t, err, errTopicNotCreated)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		topic := "topic"
		identifier1 := "identifier1"
		identifier2 := "identifier2"

		_ = messenger.CreateTopic(topic, true)
		_ = messenger.RegisterMessageProcessor(topic, identifier1, &p2pmocks.MessageProcessorStub{})
		_ = messenger.RegisterMessageProcessor(topic, identifier2, &p2pmocks.MessageProcessorStub{})

		messenger.mutOperation.RLock()
		assert.Equal(t, 2, len(messenger.topics[topic]))
		assert.NotNil(t, messenger.topics[topic][identifier1])
		assert.NotNil(t, messenger.topics[topic][identifier2])
		messenger.mutOperation.RUnlock()

		err := messenger.UnregisterMessageProcessor(topic, identifier1)
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		assert.Equal(t, 1, len(messenger.topics[topic]))
		assert.NotNil(t, messenger.topics[topic][identifier2])
		messenger.mutOperation.RUnlock()
	})
}

func TestSyncedMessenger_UnJoinAllTopics(t *testing.T) {
	t.Parallel()

	t.Run("no topics registered should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics)
		messenger.mutOperation.RUnlock()

		err := messenger.UnJoinAllTopics()
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics)
		messenger.mutOperation.RUnlock()
	})
	t.Run("one registered topic should work", func(t *testing.T) {
		t.Parallel()

		messenger, _ := NewSyncedMessenger(NewSyncedBroadcastNetwork())
		topic := "topic"
		_ = messenger.CreateTopic(topic, true)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics[topic])
		messenger.mutOperation.RUnlock()

		err := messenger.UnJoinAllTopics()
		assert.Nil(t, err)

		messenger.mutOperation.RLock()
		assert.Empty(t, messenger.topics)
		messenger.mutOperation.RUnlock()
	})
}

func TestSyncedMessenger_ConsensusMessageFilter(t *testing.T) {
	t.Parallel()

	messenger, err := NewSyncedMessenger(NewSyncedBroadcastNetwork())
	require.NoError(t, err)

	consensusTopic := common.ConsensusTopic + "_0"
	regularTopic := "regular"
	require.NoError(t, messenger.CreateTopic(consensusTopic, true))
	require.NoError(t, messenger.CreateTopic(regularTopic, true))

	var consensusCalls atomic.Uint32
	var regularCalls atomic.Uint32
	require.NoError(t, messenger.RegisterMessageProcessor(
		consensusTopic,
		"consensus",
		&p2pmocks.MessageProcessorStub{
			ProcessReceivedMessageCalled: func(_ p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
				consensusCalls.Add(1)
				return nil, nil
			},
		},
	))
	require.NoError(t, messenger.RegisterMessageProcessor(
		regularTopic,
		"regular",
		&p2pmocks.MessageProcessorStub{
			ProcessReceivedMessageCalled: func(_ p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
				regularCalls.Add(1)
				return nil, nil
			},
		},
	))

	receive := func(topic string, method p2p.BroadcastMethod) {
		messenger.receive("sender", &p2pMessage.Message{
			FromField:            []byte("sender"),
			TopicField:           topic,
			BroadcastMethodField: method,
		})
	}

	// Unknown participation preserves the full-network behavior used during START_ROUND.
	messenger.setConsensusMessageFilter(func() bool { return true })
	receive(consensusTopic, p2p.Broadcast)
	require.Equal(t, uint32(1), consensusCalls.Load())

	// A non-participant skips only consensus broadcasts.
	messenger.setConsensusMessageFilter(func() bool { return false })
	receive(consensusTopic, p2p.Broadcast)
	receive(regularTopic, p2p.Broadcast)
	receive(consensusTopic, p2p.Direct)
	require.Equal(t, uint32(2), consensusCalls.Load(), "direct consensus messages must still be delivered")
	require.Equal(t, uint32(1), regularCalls.Load(), "non-consensus broadcasts must still be delivered")
}
