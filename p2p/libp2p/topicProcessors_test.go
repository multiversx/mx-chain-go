package libp2p

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTopicProcessors(t *testing.T) {
	t.Parallel()

	tp := newTopicProcessors()

	assert.NotNil(t, tp)
}

func TestTopicProcessorsAddShouldWork(t *testing.T) {
	t.Parallel()

	tp := newTopicProcessors()

	identifier := "identifier"
	proc := &mock.MessageProcessorStub{}
	err := tp.addTopicProcessor(identifier, proc)

	assert.Nil(t, err)
	require.Equal(t, 1, len(tp.processors))
	require.Equal(t, 1, len(tp.identifiers))
	assert.True(t, proc == tp.processors[identifier]) //pointer testing
	assert.Equal(t, identifier, tp.identifiers[0])
}

func TestTopicProcessorsDoubleAddShouldErr(t *testing.T) {
	t.Parallel()

	tp := newTopicProcessors()

	identifier := "identifier"
	_ = tp.addTopicProcessor(identifier, &mock.MessageProcessorStub{})
	err := tp.addTopicProcessor(identifier, &mock.MessageProcessorStub{})

	assert.True(t, errors.Is(err, p2p.ErrMessageProcessorAlreadyDefined))
	require.Equal(t, 1, len(tp.processors))
	require.Equal(t, 1, len(tp.identifiers))
}

func TestTopicProcessorsRemoveInexistentShouldErr(t *testing.T) {
	t.Parallel()

	tp := newTopicProcessors()

	identifier := "identifier"
	err := tp.removeTopicProcessor(identifier)

	assert.True(t, errors.Is(err, p2p.ErrMessageProcessorDoesNotExists))
}

func TestTopicProcessorsRemoveShouldWork(t *testing.T) {
	t.Parallel()

	tp := newTopicProcessors()

	identifier1 := "identifier1"
	identifier2 := "identifier2"
	_ = tp.addTopicProcessor(identifier1, &mock.MessageProcessorStub{})
	_ = tp.addTopicProcessor(identifier2, &mock.MessageProcessorStub{})

	require.Equal(t, 2, len(tp.processors))
	require.Equal(t, 2, len(tp.identifiers))

	err := tp.removeTopicProcessor(identifier2)

	assert.Nil(t, err)
	require.Equal(t, 1, len(tp.processors))
	require.Equal(t, 1, len(tp.identifiers))

	err = tp.removeTopicProcessor(identifier1)

	assert.Nil(t, err)
	require.Equal(t, 0, len(tp.processors))
	require.Equal(t, 0, len(tp.identifiers))
}

func TestTopicProcessorsGetListShouldWorkAndPreserveOrder(t *testing.T) {
	t.Parallel()

	tp := newTopicProcessors()

	identifier1 := "identifier1"
	identifier2 := "identifier2"
	identifier3 := "identifier3"
	handler1 := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return nil
		},
	}
	handler2 := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return nil
		},
	}
	handler3 := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return nil
		},
	}

	_ = tp.addTopicProcessor(identifier3, handler3)
	_ = tp.addTopicProcessor(identifier1, handler1)
	_ = tp.addTopicProcessor(identifier2, handler2)

	require.Equal(t, 3, len(tp.processors))
	require.Equal(t, 3, len(tp.identifiers))

	identifiers, handlers := tp.getList()
	assert.Equal(t, identifiers, []string{identifier3, identifier1, identifier2})
	assert.Equal(t, handlers, []p2p.MessageProcessor{handler3, handler1, handler2})

	_ = tp.removeTopicProcessor(identifier1)
	identifiers, handlers = tp.getList()
	assert.Equal(t, identifiers, []string{identifier3, identifier2})
	assert.Equal(t, handlers, []p2p.MessageProcessor{handler3, handler2})

	_ = tp.removeTopicProcessor(identifier2)
	identifiers, handlers = tp.getList()
	assert.Equal(t, identifiers, []string{identifier3})
	assert.Equal(t, handlers, []p2p.MessageProcessor{handler3})

	_ = tp.removeTopicProcessor(identifier3)
	identifiers, handlers = tp.getList()
	assert.Equal(t, identifiers, make([]string, 0))
	assert.Equal(t, handlers, make([]p2p.MessageProcessor, 0))
}
