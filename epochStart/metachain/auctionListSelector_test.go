package metachain

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
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

func createFullAuctionListSelectorArgs(config []config.MaxNodesChangeConfig) (AuctionListSelectorArgs, ArgsNewEpochStartSystemSCProcessing) {
	epochNotifier := forking.NewGenericEpochNotifier()
	nodesConfigProvider, _ := notifier.NewNodesConfigProvider(epochNotifier, config)

	argsSystemSC, _ := createFullArgumentsForSystemSCProcessing(0, createMemUnit())
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)
	return AuctionListSelectorArgs{
		ShardCoordinator:             shardCoordinator,
		StakingDataProvider:          argsSystemSC.StakingDataProvider,
		MaxNodesChangeConfigProvider: nodesConfigProvider,
	}, argsSystemSC
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

	args, argsSystemSC := createFullAuctionListSelectorArgs([]config.MaxNodesChangeConfig{{MaxNumNodes: 1}})
	owner1 := []byte("owner1")
	owner2 := []byte("owner2")

	owner1StakedKeys := [][]byte{[]byte("pubKey0")}
	owner2StakedKeys := [][]byte{[]byte("pubKey1")}

	validatorsInfo := state.NewShardValidatorsInfoMap()
	_ = validatorsInfo.Add(createValidatorInfo(owner1StakedKeys[0], common.EligibleList, owner1, 0))
	_ = validatorsInfo.Add(createValidatorInfo(owner2StakedKeys[0], common.AuctionList, owner2, 0))

	stakingcommon.RegisterValidatorKeys(argsSystemSC.UserAccountsDB, owner1, owner1, owner1StakedKeys, big.NewInt(1000), argsSystemSC.Marshalizer)
	stakingcommon.RegisterValidatorKeys(argsSystemSC.UserAccountsDB, owner2, owner2, owner2StakedKeys, big.NewInt(1000), argsSystemSC.Marshalizer)

	err := args.StakingDataProvider.FillValidatorInfo(owner1StakedKeys[0])
	require.Nil(t, err)
	err = args.StakingDataProvider.FillValidatorInfo(owner2StakedKeys[0])
	require.Nil(t, err)

	als, _ := NewAuctionListSelector(args)
	err = als.SelectNodesFromAuctionList(validatorsInfo, nil, []byte("rnd"))
	require.Nil(t, err)

	expectedValidatorsInfo := map[uint32][]state.ValidatorInfoHandler{
		0: {
			createValidatorInfo(owner1StakedKeys[0], common.EligibleList, owner1, 0),
			createValidatorInfo(owner2StakedKeys[0], common.AuctionList, owner2, 0),
		},
	}
	require.Equal(t, expectedValidatorsInfo, validatorsInfo.GetShardValidatorsInfoMap())
}

//TODO: probably remove this test
/*
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

	err := als.SelectNodesFromAuctionList(validatorsInfo, nil, []byte("rnd"))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errGetNodeTopUp.Error()))
	require.True(t, strings.Contains(err.Error(), epochStart.ErrSortAuctionList.Error()))
}
*/
