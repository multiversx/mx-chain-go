package topicsender

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockArgTopicRequestSender() ArgTopicRequestSender {
	return ArgTopicRequestSender{
		ArgBaseTopicSender: ArgBaseTopicSender{
			Messenger:         &mock.MessageHandlerStub{},
			TopicName:         "topic",
			OutputAntiflooder: &mock.P2PAntifloodHandlerStub{},
			PreferredPeersHolder: &p2pmocks.PeersHolderStub{
				GetCalled: func() map[uint32][]core.PeerID {
					return map[uint32][]core.PeerID{}
				},
			},
			TargetShardId: 0,
		},
		Marshaller:                  &mock.MarshalizerMock{},
		Randomizer:                  &mock.IntRandomizerStub{},
		PeerListCreator:             &mock.PeerListCreatorStub{},
		NumIntraShardPeers:          2,
		NumCrossShardPeers:          2,
		NumFullHistoryPeers:         2,
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		SelfShardIdProvider:         mock.NewMultipleShardsCoordinatorMock(),
		PeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
	}
}

func TestNewTopicRequestSender(t *testing.T) {
	t.Parallel()

	t.Run("", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.PeerListCreator = nil
		trs, err := NewTopicRequestSender(arg)

		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilPeerListCreator, err)
	})
}
