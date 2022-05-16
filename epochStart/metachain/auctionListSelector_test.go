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
