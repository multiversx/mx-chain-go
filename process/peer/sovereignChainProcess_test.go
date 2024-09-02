package peer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignChainValidatorStatisticsProcessor_ShouldErrNilValidatorStatistics(t *testing.T) {
	t.Parallel()

	scvs, err := peer.NewSovereignChainValidatorStatisticsProcessor(nil)
	assert.Nil(t, scvs)
	assert.Equal(t, process.ErrNilValidatorStatistics, err)
}

func TestNewSovereignChainValidatorStatisticsProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	vs, _ := peer.NewValidatorStatisticsProcessor(args)

	scvs, err := peer.NewSovereignChainValidatorStatisticsProcessor(vs)
	assert.NotNil(t, scvs)
	assert.Nil(t, err)
}

func TestSovereignValidatorStatisticsProcessor_UpdateShardDataPeerState_IncreasesConsensusCurrentShardBlock_SameEpoch(t *testing.T) {
	t.Parallel()

	consensusGroup := make(map[string][]nodesCoordinator.Validator)

	cache := createMockCache()
	prevHeader, header := generateTestShardBlockHeaders(cache)

	header.Round = prevHeader.Round + 1
	header.Nonce = prevHeader.Nonce + 1
	header.Epoch = 1
	header.PrevRandSeed = prevHeader.RandSeed
	sovHdr := &block.SovereignChainHeader{
		Header: header,
	}

	v1 := shardingMocks.NewValidatorMock([]byte("pk1"), 1, 1)
	v2 := shardingMocks.NewValidatorMock([]byte("pk2"), 1, 1)

	currentHeaderConsensus := []nodesCoordinator.Validator{v1, v2}
	consensusGroup["prevRandSeed_1_0_1"] = currentHeaderConsensus

	arguments := createUpdateTestArgs(consensusGroup)
	vs, _ := peer.NewValidatorStatisticsProcessor(arguments)
	validatorStatistics, _ := peer.NewSovereignChainValidatorStatisticsProcessor(vs)

	prevHash := string(header.GetPrevHash())
	_, err := validatorStatistics.UpdatePeerState(sovHdr, map[string]data.CommonHeaderHandler{
		prevHash: cache[prevHash],
	})
	require.Nil(t, err)

	pa1, _ := validatorStatistics.LoadPeerAccount(v1.PubKey())
	leader := pa1.(*stateMock.PeerAccountHandlerMock)
	pa2, _ := validatorStatistics.LoadPeerAccount(v2.PubKey())
	validator := pa2.(*stateMock.PeerAccountHandlerMock)

	require.Equal(t, uint32(1), leader.IncreaseLeaderSuccessRateValue)
	require.Equal(t, uint32(1), validator.IncreaseValidatorSuccessRateValue)
}
