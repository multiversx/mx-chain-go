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

func TestNewP2PAntiflood_NilFloodPreventerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(nil, &mock.TopicAntiFloodStub{})
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, p2p.ErrNilFloodPreventer))
}

func TestNewP2PAntiflood_NilTopicFloodPreventerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{}, nil)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, p2p.ErrNilTopicFloodPreventer))
}

func TestNewP2PAntiflood_ShouldWork(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{}, &mock.TopicAntiFloodStub{})

	assert.False(t, check.IfNil(afm))
	assert.Nil(t, err)
}

func TestP2PAntiflood_SettingInnerFloodPreventerToNil(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{}, &mock.TopicAntiFloodStub{})

	afm.FloodPreventer = nil
	assert.True(t, check.IfNil(afm))
}

//------- CanProcessMessage

func TestP2PAntiflood_CanProcessMessageNilFloodPreventerShouldError(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{}, &mock.TopicAntiFloodStub{})
	afm.FloodPreventer = nil

	err := afm.CanProcessMessage(&mock.P2PMessageMock{}, "connected peer")
	assert.Equal(t, p2p.ErrNilFloodPreventer, err)
}

func TestP2PAntiflood_CanProcessMessageNilMessageShouldError(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{}, &mock.TopicAntiFloodStub{})

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
		&mock.FloodPreventerStub{
			IncreaseLoadGlobalCalled: func(identifier string, size uint64) error {
				if identifier != fromConnectedPeer.Pretty() {
					assert.Fail(t, "should have been the connected peer")
				}

				return process.ErrSystemBusy
			},
		},
		&mock.TopicAntiFloodStub{},
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
	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{
		IncreaseLoadGlobalCalled: func(identifier string, size uint64) error {
			if identifier != fromConnectedPeer.Pretty() {
				return process.ErrSystemBusy
			}

			return nil
		},
		IncreaseLoadCalled: func(identifier string, size uint64) error {
			if identifier == message.PeerField.Pretty() {
				return process.ErrSystemBusy
			}

			return nil
		},
	},
		&mock.TopicAntiFloodStub{},
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
	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{
		IncreaseLoadGlobalCalled: func(identifier string, size uint64) error {
			return nil
		},
		IncreaseLoadCalled: func(identifier string, size uint64) error {
			return nil
		},
	},
		&mock.TopicAntiFloodStub{},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.Nil(t, err)
}

//------- CanProcessMessagesOnTopic

func TestP2pAntiflood_CanProcessMessagesOnTopicCanNotAccumulateShouldError(t *testing.T) {
	t.Parallel()

	numMessagesCall := uint32(78)
	topicCall := "topic"
	identifierCall := p2p.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.FloodPreventerStub{},
		&mock.TopicAntiFloodStub{
			IncreaseLoadCalled: func(identifier string, topic string, numMessages uint32) error {
				if identifier == identifierCall.Pretty() && topic == topicCall && numMessages == numMessagesCall {
					return process.ErrSystemBusy
				}

				return nil
			},
		},
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
		&mock.FloodPreventerStub{},
		&mock.TopicAntiFloodStub{
			IncreaseLoadCalled: func(identifier string, topic string, numMessages uint32) error {
				if identifier == identifierCall.Pretty() && topic == topicCall && numMessages == numMessagesCall {
					return nil
				}

				return process.ErrSystemBusy
			},
		},
	)

	err := afm.CanProcessMessagesOnTopic(identifierCall, topicCall, numMessagesCall)

	assert.Nil(t, err)
}
