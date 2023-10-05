package chainSimulator

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, messenger.AddPeerTopicNotifier(nil))
	assert.Zero(t, messenger.Port())
	assert.Nil(t, messenger.SetPeerDenialEvaluator(nil))
	assert.Nil(t, messenger.SetThresholdMinConnectedPeers(0))
	assert.Zero(t, messenger.ThresholdMinConnectedPeers())
	assert.True(t, messenger.IsConnectedToTheNetwork())
	assert.Nil(t, messenger.SetPeerShardResolver(nil))
	assert.Nil(t, messenger.ConnectToPeer(""))
	assert.Nil(t, messenger.Bootstrap())
	assert.Nil(t, messenger.ProcessReceivedMessage(nil, "", nil))

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
