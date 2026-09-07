package resolvers

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fromConnectedPeer = core.PeerID("from connected peer")

//------- canProcessMessage

func TestMessageProcessor_CanProcessNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mp := &messageProcessor{}

	err := mp.canProcessMessage(nil, "")

	assert.True(t, errors.Is(err, dataRetriever.ErrNilMessage))
}

func TestMessageProcessor_CanProcessErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return expectedErr
			},
		},
	}

	err := mp.canProcessMessage(&p2pmocks.P2PMessageMock{}, "")

	assert.True(t, errors.Is(err, expectedErr))
}

func TestMessageProcessor_CanProcessOnTopicErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
				return expectedErr
			},
		},
	}

	err := mp.canProcessMessage(&p2pmocks.P2PMessageMock{}, "")

	assert.True(t, errors.Is(err, expectedErr))
}

func TestMessageProcessor_CanProcessThrottlerNotAllowingShouldErr(t *testing.T) {
	t.Parallel()

	canProcessWasCalled := false
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
				return nil
			},
		},
		throttler: &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				canProcessWasCalled = true
				return false
			},
		},
	}

	err := mp.canProcessMessage(&p2pmocks.P2PMessageMock{}, "")

	assert.True(t, errors.Is(err, dataRetriever.ErrSystemBusy))
	assert.True(t, canProcessWasCalled)
}

func TestMessageProcessor_CanProcessShouldWork(t *testing.T) {
	t.Parallel()

	canProcessWasCalled := false
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
				return nil
			},
		},
		throttler: &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				canProcessWasCalled = true
				return true
			},
		},
	}

	err := mp.canProcessMessage(&p2pmocks.P2PMessageMock{}, "")

	assert.Nil(t, err)
	assert.True(t, canProcessWasCalled)
}

//------- parseReceivedMessage

func TestMessageProcessor_ParseReceivedMessageMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	originatorPid := core.PeerID("originator")
	originatorBlackListed := false
	fromConnectedPeerBlackListed := false
	var mp = &messageProcessor{
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		},
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			BlacklistPeerCalled: func(peer core.PeerID, reason string, duration time.Duration) {
				if peer == originatorPid {
					originatorBlackListed = true
				}
				if peer == fromConnectedPeer {
					fromConnectedPeerBlackListed = true
				}
			},
		},
	}
	msg := &p2pmocks.P2PMessageMock{
		DataField: make([]byte, 0),
		PeerField: originatorPid,
	}
	rd, err := mp.parseReceivedMessage(msg, fromConnectedPeer)

	assert.True(t, originatorBlackListed)
	assert.True(t, fromConnectedPeerBlackListed)
	assert.Equal(t, err, expectedErr)
	assert.Nil(t, rd)
}

func TestMessageProcessor_ParseReceivedMessageNilValueFieldShouldErr(t *testing.T) {
	t.Parallel()

	mp := &messageProcessor{
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
	}

	msg := &p2pmocks.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	rd, err := mp.parseReceivedMessage(msg, fromConnectedPeer)

	assert.Equal(t, err, dataRetriever.ErrNilValue)
	assert.Nil(t, rd)
}

func TestMessageProcessor_ParseReceivedMessageShouldWork(t *testing.T) {
	t.Parallel()

	expectedValue := []byte("expected value")
	mp := &messageProcessor{
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				rd := obj.(*dataRetriever.RequestData)
				rd.Value = expectedValue

				return nil
			},
		},
	}

	msg := &p2pmocks.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	rd, err := mp.parseReceivedMessage(msg, fromConnectedPeer)

	assert.Nil(t, err)
	require.NotNil(t, rd)
	assert.Equal(t, expectedValue, rd.Value)
}

func TestMessageProcessor_ParseRequestedHashesTooManyShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hashes := make([][]byte, maxHashesInRequest+1)
	for index := range hashes {
		hashes[index] = []byte{byte(index)}
	}
	hashesBuff, err := marshalizer.Marshal(&batch.Batch{Data: hashes})
	require.NoError(t, err)

	wasAccounted := false
	mp := &messageProcessor{
		marshalizer: marshalizer,
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessagesOnTopicCalled: func(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte) error {
				wasAccounted = true
				return nil
			},
		},
	}

	result, err := mp.parseRequestedHashes(hashesBuff, fromConnectedPeer, nil)

	require.ErrorIs(t, err, dataRetriever.ErrBadRequest)
	require.Nil(t, result)
	require.False(t, wasAccounted)
}

func TestMessageProcessor_ParseRequestedHashesShouldAccountAndDeduplicate(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hashes := [][]byte{[]byte("hash1"), []byte("hash1"), []byte("hash2")}
	hashesBuff, err := marshalizer.Marshal(&batch.Batch{Data: hashes})
	require.NoError(t, err)

	topic := "topic"
	sequence := []byte("sequence")
	mp := &messageProcessor{
		marshalizer: marshalizer,
		topic:       topic,
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, receivedTopic string, numMessages uint32, totalSize uint64, receivedSequence []byte) error {
				require.Equal(t, fromConnectedPeer, peer)
				require.Equal(t, topic, receivedTopic)
				require.Equal(t, uint32(len(hashes)), numMessages)
				require.Equal(t, uint64(len(hashesBuff)), totalSize)
				require.Equal(t, sequence, receivedSequence)

				return nil
			},
		},
	}

	result, err := mp.parseRequestedHashes(hashesBuff, fromConnectedPeer, sequence)

	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("hash1"), []byte("hash2")}, result)
}

func TestMessageProcessor_ParseRequestedHashesAntifloodErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	marshalizer := &mock.MarshalizerMock{}
	hashesBuff, err := marshalizer.Marshal(&batch.Batch{Data: [][]byte{[]byte("hash")}})
	require.NoError(t, err)

	mp := &messageProcessor{
		marshalizer: marshalizer,
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessagesOnTopicCalled: func(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte) error {
				return expectedErr
			},
		},
	}

	result, err := mp.parseRequestedHashes(hashesBuff, fromConnectedPeer, nil)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)
}

func TestHashArrayReplyShouldLimitAccumulatedData(t *testing.T) {
	t.Parallel()

	reply := newHashArrayReply()
	first := bytes.Repeat([]byte{1}, maxHashArrayReplySize/2+1)
	second := bytes.Repeat([]byte{2}, maxHashArrayReplySize/2)

	require.True(t, reply.add(first))
	require.False(t, reply.add(second))
	require.Equal(t, [][]byte{first}, reply.data)
	require.Equal(t, len(first), reply.size)
}

func TestHashArrayReplyShouldAllowOneLargeItem(t *testing.T) {
	t.Parallel()

	reply := newHashArrayReply()
	largeItem := bytes.Repeat([]byte{1}, maxHashArrayReplySize+1)

	require.True(t, reply.add(largeItem))
	require.False(t, reply.add([]byte{2}))
	require.Equal(t, [][]byte{largeItem}, reply.data)
}

func TestNewHashArrayReplyShouldHaveNonNilData(t *testing.T) {
	t.Parallel()

	reply := newHashArrayReply()

	require.NotNil(t, reply.data)
}
