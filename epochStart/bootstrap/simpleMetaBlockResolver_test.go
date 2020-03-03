package bootstrap_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	mock2 "github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/mock"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/require"
)

func TestNewSimpleMetaBlocksResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	smbr, err := bootstrap.NewSimpleMetaBlocksResolver(nil, &mock.MarshalizerMock{})
	require.Nil(t, smbr)
	require.Equal(t, bootstrap.ErrNilMessenger, err)
}

func TestNewSimpleMetaBlocksResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	smbr, err := bootstrap.NewSimpleMetaBlocksResolver(&mock2.MessengerStub{}, nil)
	require.Nil(t, smbr)
	require.Equal(t, bootstrap.ErrNilMarshalizer, err)
}

func TestNewSimpleMetaBlocksResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	smbr, err := bootstrap.NewSimpleMetaBlocksResolver(&mock2.MessengerStub{}, &mock.MarshalizerMock{})
	require.Nil(t, err)
	require.False(t, check.IfNil(smbr))
}

func TestSimpleMetaBlocksResolver_RequestEpochStartMetaBlock(t *testing.T) {
	t.Parallel()

	requestWasSent := false

	messenger := &mock2.MessengerStub{
		ConnectedPeersOnTopicCalled: func(_ string) []p2p.PeerID {
			return []p2p.PeerID{"peer1", "peer2", "peer2"}
		},
		SendToConnectedPeerCalled: func(_ string, _ []byte, _ p2p.PeerID) error {
			requestWasSent = true
			return nil
		},
	}
	smbr, _ := bootstrap.NewSimpleMetaBlocksResolver(messenger, &mock.MarshalizerMock{})

	err := smbr.RequestEpochStartMetaBlock(0)
	require.Nil(t, err)
	require.True(t, requestWasSent)
}
