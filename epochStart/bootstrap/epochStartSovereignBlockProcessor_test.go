package bootstrap

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func createSovEpochStartBlockProcessor() *epochStartSovereignBlockProcessor {
	esmbp, _ := NewEpochStartMetaBlockProcessor(
		&p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"peer_0", "peer_1"}
			},
		},
		&testscommon.RequestHandlerStub{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		99,
		2,
		2,
	)
	return newEpochStartSovereignBlockProcessor(esmbp)
}

func TestEpochStartSovereignBlockProcessor_setNumPeers(t *testing.T) {
	t.Parallel()

	numIntra := 4
	numCross := 3
	wasNumPeersSet := false
	requestHandler := &testscommon.RequestHandlerStub{
		SetNumPeersToQueryCalled: func(key string, intra int, cross int) error {
			require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), key)
			require.Equal(t, numIntra, intra)
			require.Zero(t, cross)

			wasNumPeersSet = true
			return nil
		},
	}
	sovProc := createSovEpochStartBlockProcessor()
	err := sovProc.setNumPeers(requestHandler, numIntra, numCross)
	require.Nil(t, err)
	require.True(t, wasNumPeersSet)
}

func TestEpochStartSovereignBlockProcessor_getRequestTopic(t *testing.T) {
	t.Parallel()

	sovProc := createSovEpochStartBlockProcessor()
	require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), sovProc.getTopic())
}
