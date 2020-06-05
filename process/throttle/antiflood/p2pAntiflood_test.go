package antiflood_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/stretchr/testify/assert"
)

func TestNewP2PAntiflood_NilBlacklistHandlerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		nil,
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, process.ErrNilBlackListHandler))
}

func TestNewP2PAntiflood_EmptyFloodPreventerListShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
	)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, process.ErrEmptyFloodPreventerList))
}

func TestNewP2PAntiflood_NilTopicFloodPreventerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		nil,
		&mock.FloodPreventerStub{},
	)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, process.ErrNilTopicFloodPreventer))
}

func TestNewP2PAntiflood_ShouldWork(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	assert.False(t, check.IfNil(afm))
	assert.Nil(t, err)
}

//------- CanProcessMessage

func TestP2PAntiflood_CanProcessMessageNilMessageShouldError(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	err := afm.CanProcessMessage(nil, "connected peer")
	assert.Equal(t, p2p.ErrNilMessage, err)
}

func TestP2PAntiflood_CanNotIncrementFromConnectedPeerShouldError(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := p2p.PeerID("from connected peer")
	message := &mock.P2PMessageMock{
		DataField: []byte("data"),
		FromField: messageOriginator,
	}
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(identifier string, size uint64) error {
				if identifier != fromConnectedPeer.Pretty() {
					assert.Fail(t, "should have been the connected peer")
				}

				return process.ErrSystemBusy
			},
		},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestP2PAntiflood_CanNotIncrementMessageOriginatorShouldError(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := p2p.PeerID("from connected peer")
	message := &mock.P2PMessageMock{
		DataField: []byte("data"),
		FromField: messageOriginator,
		PeerField: p2p.PeerID(messageOriginator),
	}
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(identifier string, size uint64) error {
				if identifier == message.PeerField.Pretty() {
					return process.ErrSystemBusy
				}
				if identifier != fromConnectedPeer.Pretty() {
					return process.ErrSystemBusy
				}

				return nil
			},
		},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestP2PAntiflood_ShouldWork(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := p2p.PeerID("from connected peer")
	message := &mock.P2PMessageMock{
		DataField: []byte("data"),
		PeerField: p2p.PeerID(messageOriginator),
	}
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(identifier string, size uint64) error {
				return nil
			},
		},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.Nil(t, err)
}

func TestP2PAntiflood_ShouldWorkWithMoreThanOneFlodPreventer(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := p2p.PeerID("from connected peer")
	message := &mock.P2PMessageMock{
		DataField: []byte("data"),
		PeerField: p2p.PeerID(messageOriginator),
	}
	numIncreasedLoads := 0

	fp := &mock.FloodPreventerStub{
		IncreaseLoadCalled: func(identifier string, size uint64) error {
			numIncreasedLoads++
			return nil
		},
	}

	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		fp,
		fp,
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.Nil(t, err)
	assert.Equal(t, 4, numIncreasedLoads)
}

//------- CanProcessMessagesOnTopic

func TestP2pAntiflood_CanProcessMessagesOnTopicCanNotAccumulateShouldError(t *testing.T) {
	t.Parallel()

	numMessagesCall := uint32(78)
	topicCall := "topic"
	identifierCall := p2p.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{
			IncreaseLoadCalled: func(identifier string, topic string, numMessages uint32) error {
				if identifier == identifierCall.Pretty() && topic == topicCall && numMessages == numMessagesCall {
					return process.ErrSystemBusy
				}

				return nil
			},
		},
		&mock.FloodPreventerStub{},
	)

	err := afm.CanProcessMessagesOnTopic(identifierCall, topicCall, numMessagesCall)

	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestP2pAntiflood_CanProcessMessagesOnTopicCanAccumulateShouldWork(t *testing.T) {
	t.Parallel()

	numMessagesCall := uint32(78)
	topicCall := "topic"
	identifierCall := p2p.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{
			IncreaseLoadCalled: func(identifier string, topic string, numMessages uint32) error {
				if identifier == identifierCall.Pretty() && topic == topicCall && numMessages == numMessagesCall {
					return nil
				}

				return process.ErrSystemBusy
			},
		},
		&mock.FloodPreventerStub{},
	)

	err := afm.CanProcessMessagesOnTopic(identifierCall, topicCall, numMessagesCall)

	assert.Nil(t, err)
}

func TestP2pAntiflood_CanProcessMessagesOriginatorIsBlacklistedShouldErr(t *testing.T) {
	t.Parallel()

	identifier := p2p.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(identifier string, size uint64) error {
				return nil
			},
		},
	)
	message := &mock.P2PMessageMock{
		DataField: []byte("data"),
		PeerField: identifier,
	}

	err := afm.CanProcessMessage(message, identifier)

	assert.True(t, errors.Is(err, process.ErrOriginatorIsBlacklisted))
}

func TestP2pAntiflood_ResetForTopicSetMaxMessagesShouldWork(t *testing.T) {
	t.Parallel()

	resetTopicCalled := false
	resetTopicParameter := ""
	setMaxMessagesForTopicCalled := false
	setMaxMessagesForTopicParameter1 := ""
	setMaxMessagesForTopicParameter2 := uint32(0)
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{
			ResetForTopicCalled: func(topic string) {
				resetTopicCalled = true
				resetTopicParameter = topic
			},
			SetMaxMessagesForTopicCalled: func(topic string, num uint32) {
				setMaxMessagesForTopicCalled = true
				setMaxMessagesForTopicParameter1 = topic
				setMaxMessagesForTopicParameter2 = num
			},
		},
		&mock.FloodPreventerStub{},
	)

	resetTopic := "reset topic"
	afm.ResetForTopic(resetTopic)
	assert.True(t, resetTopicCalled)
	assert.Equal(t, resetTopic, resetTopicParameter)

	setMaxMessagesForTopic := "set max message for topic"
	setMaxMessagesForTopicNum := uint32(77463)
	afm.SetMaxMessagesForTopic(setMaxMessagesForTopic, setMaxMessagesForTopicNum)
	assert.True(t, setMaxMessagesForTopicCalled)
	assert.Equal(t, setMaxMessagesForTopic, setMaxMessagesForTopicParameter1)
	assert.Equal(t, setMaxMessagesForTopicNum, setMaxMessagesForTopicParameter2)
}

func TestP2pAntiflood_ApplyConsensusSize(t *testing.T) {
	t.Parallel()

	wasCalled := false
	expectedSize := 878264
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.BlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			ApplyConsensusSizeCalled: func(size int) {
				assert.Equal(t, expectedSize, size)
				wasCalled = true
			},
		},
	)

	afm.ApplyConsensusSize(expectedSize)
	assert.True(t, wasCalled)
}
