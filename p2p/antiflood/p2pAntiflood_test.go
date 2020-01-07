package antiflood_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/antiflood"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewP2PAntiflood_NilFloodPreventerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(nil)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, p2p.ErrNilFloodPreventer))
}

func TestNewP2PAntiflood_ShouldWork(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{})

	assert.False(t, check.IfNil(afm))
	assert.Nil(t, err)
}

func TestP2PAntiflood_SettingInnerFloodPreventerToNil(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{})

	afm.FloodPreventer = nil
	assert.True(t, check.IfNil(afm))
}

//------- CanProcessMessage

func TestP2PAntiflood_CanProcessMessageNilFloodPreventerShouldError(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{})
	afm.FloodPreventer = nil

	err := afm.CanProcessMessage(&mock.P2PMessageMock{}, "connected peer")
	assert.Equal(t, p2p.ErrNilFloodPreventer, err)
}

func TestP2PAntiflood_CanProcessMessageNilMessageShouldError(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{})

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
	afm, _ := antiflood.NewP2PAntiflood(&mock.FloodPreventerStub{
		AccumulateGlobalCalled: func(identifier string, size uint64) bool {
			if identifier != fromConnectedPeer.Pretty() {
				assert.Fail(t, "should have been the connected peer")
			}

			return false
		},
	})

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.True(t, errors.Is(err, p2p.ErrSystemBusy))
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
		AccumulateGlobalCalled: func(identifier string, size uint64) bool {
			return identifier == fromConnectedPeer.Pretty()
		},
		AccumulateCalled: func(identifier string, size uint64) bool {
			return identifier != message.PeerField.Pretty()
		},
	})

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.True(t, errors.Is(err, p2p.ErrSystemBusy))
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
		AccumulateGlobalCalled: func(identifier string, size uint64) bool {
			return true
		},
		AccumulateCalled: func(identifier string, size uint64) bool {
			return true
		},
	})

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.Nil(t, err)
}
