package metachain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func createAuctionListSelectorArgs() AuctionListSelectorArgs {
	epochNotifier := forking.NewGenericEpochNotifier()
	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(epochNotifier, nil)

	argsStakingDataProvider := createStakingDataProviderArgs()
	stakingSCProvider, _ := NewStakingDataProvider(argsStakingDataProvider)

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)
	return AuctionListSelectorArgs{
		ShardCoordinator:             shardCoordinator,
		StakingDataProvider:          stakingSCProvider,
		MaxNodesChangeConfigProvider: nodesConfigProvider,
	}
}

func TestNewAuctionListSelector(t *testing.T) {
	t.Parallel()

	t.Run("nil shard coordinator", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs()
		args.ShardCoordinator = nil
		als, err := NewAuctionListSelector(args)
		require.Nil(t, als)
		require.Equal(t, epochStart.ErrNilShardCoordinator, err)
	})

	t.Run("nil staking data provider", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs()
		args.StakingDataProvider = nil
		als, err := NewAuctionListSelector(args)
		require.Nil(t, als)
		require.Equal(t, epochStart.ErrNilStakingDataProvider, err)
	})

	t.Run("nil max nodes change config provider", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs()
		args.MaxNodesChangeConfigProvider = nil
		als, err := NewAuctionListSelector(args)
		require.Nil(t, als)
		require.Equal(t, epochStart.ErrNilMaxNodesChangeConfigProvider, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs()
		als, err := NewAuctionListSelector(args)
		require.NotNil(t, als)
		require.Nil(t, err)
	})
}

/*
func TestAuctionListSelector_EpochConfirmedCorrectMaxNumNodesAfterNodeRestart(t *testing.T) {
	t.Parallel()

	args := createAuctionListSelectorArgs()
	nodesConfigEpoch0 := config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            36,
		NodesToShufflePerShard: 4,
	}
	nodesConfigEpoch1 := config.MaxNodesChangeConfig{
		EpochEnable:            1,
		MaxNumNodes:            56,
		NodesToShufflePerShard: 2,
	}
	nodesConfigEpoch6 := config.MaxNodesChangeConfig{
		EpochEnable:            6,
		MaxNumNodes:            48,
		NodesToShufflePerShard: 1,
	}

	args.MaxNodesEnableConfig = []config.MaxNodesChangeConfig{
		nodesConfigEpoch0,
		nodesConfigEpoch1,
		nodesConfigEpoch6,
	}

	als, _ := NewAuctionListSelector(args)

	als.EpochConfirmed(0, 0)
	require.Equal(t, nodesConfigEpoch0, als.currentNodesEnableConfig)

	als.EpochConfirmed(1, 1)
	require.Equal(t, nodesConfigEpoch1, als.currentNodesEnableConfig)

	for epoch := uint32(2); epoch <= 5; epoch++ {
		als.EpochConfirmed(epoch, uint64(epoch))
		require.Equal(t, nodesConfigEpoch1, als.currentNodesEnableConfig)
	}

	// simulate restart
	als.EpochConfirmed(0, 0)
	als.EpochConfirmed(5, 5)
	require.Equal(t, nodesConfigEpoch1, als.currentNodesEnableConfig)

	als.EpochConfirmed(6, 6)
	require.Equal(t, nodesConfigEpoch6, als.currentNodesEnableConfig)

	// simulate restart
	als.EpochConfirmed(0, 0)
	als.EpochConfirmed(6, 6)
	require.Equal(t, nodesConfigEpoch6, als.currentNodesEnableConfig)

	for epoch := uint32(7); epoch <= 20; epoch++ {
		als.EpochConfirmed(epoch, uint64(epoch))
		require.Equal(t, nodesConfigEpoch6, als.currentNodesEnableConfig)
	}

	// simulate restart
	als.EpochConfirmed(1, 1)
	als.EpochConfirmed(21, 21)
	require.Equal(t, nodesConfigEpoch6, als.currentNodesEnableConfig)
}

*/
