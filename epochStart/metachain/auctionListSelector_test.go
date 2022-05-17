package metachain

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/require"
)

func createAuctionListSelectorArgs(config []config.MaxNodesChangeConfig) AuctionListSelectorArgs {
	epochNotifier := forking.NewGenericEpochNotifier()
	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(epochNotifier, config)

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
		args := createAuctionListSelectorArgs(nil)
		args.ShardCoordinator = nil
		als, err := NewAuctionListSelector(args)
		require.Nil(t, als)
		require.Equal(t, epochStart.ErrNilShardCoordinator, err)
	})

	t.Run("nil staking data provider", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs(nil)
		args.StakingDataProvider = nil
		als, err := NewAuctionListSelector(args)
		require.Nil(t, als)
		require.Equal(t, epochStart.ErrNilStakingDataProvider, err)
	})

	t.Run("nil max nodes change config provider", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs(nil)
		args.MaxNodesChangeConfigProvider = nil
		als, err := NewAuctionListSelector(args)
		require.Nil(t, als)
		require.Equal(t, epochStart.ErrNilMaxNodesChangeConfigProvider, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createAuctionListSelectorArgs(nil)
		als, err := NewAuctionListSelector(args)
		require.NotNil(t, als)
		require.Nil(t, err)
	})
}

func TestAuctionListSelector_SelectNodesFromAuctionListNotEnoughSlotsForAuctionNodes(t *testing.T) {
	t.Parallel()

	args := createAuctionListSelectorArgs([]config.MaxNodesChangeConfig{{MaxNumNodes: 1}})
	als, _ := NewAuctionListSelector(args)

	owner1 := []byte("owner1")
	owner2 := []byte("owner2")

	owner1StakedKeys := [][]byte{[]byte("pubKey0")}
	owner2StakedKeys := [][]byte{[]byte("pubKey1")}

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(createValidatorInfo(owner1StakedKeys[0], common.EligibleList, owner1, 0))
	_ = validatorsInfo.Add(createValidatorInfo(owner2StakedKeys[0], common.AuctionList, owner2, 0))

	err := als.SelectNodesFromAuctionList(validatorsInfo, []byte("rnd"))
	require.Nil(t, err)

	expectedValidatorsInfo := map[uint32][]state.ValidatorInfoHandler{
		0: {
			createValidatorInfo(owner1StakedKeys[0], common.EligibleList, owner1, 0),
			createValidatorInfo(owner2StakedKeys[0], common.AuctionList, owner2, 0),
		},
	}
	require.Equal(t, expectedValidatorsInfo, validatorsInfo.GetShardValidatorsInfoMap())
}

func TestSystemSCProcessor_ProcessSystemSmartContractStakingV4EnabledErrSortingAuctionList(t *testing.T) {
	t.Parallel()

	args := createAuctionListSelectorArgs([]config.MaxNodesChangeConfig{{MaxNumNodes: 10}})

	errGetNodeTopUp := errors.New("error getting top up per node")
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			switch string(blsKey) {
			case "pubKey0", "pubKey1":
				return nil, errGetNodeTopUp
			default:
				require.Fail(t, "should not call this func with other params")
				return nil, nil
			}
		},
	}
	als, _ := NewAuctionListSelector(args)

	owner := []byte("owner")
	ownerStakedKeys := [][]byte{[]byte("pubKey0"), []byte("pubKey1")}

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(createValidatorInfo(ownerStakedKeys[0], common.AuctionList, owner, 0))
	_ = validatorsInfo.Add(createValidatorInfo(ownerStakedKeys[1], common.AuctionList, owner, 0))

	err := als.SelectNodesFromAuctionList(validatorsInfo, []byte("rnd"))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errGetNodeTopUp.Error()))
	require.True(t, strings.Contains(err.Error(), epochStart.ErrSortAuctionList.Error()))
}
