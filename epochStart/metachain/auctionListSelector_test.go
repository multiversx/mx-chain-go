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
	argsSystemSC.MaxNodesChangeConfigProvider = nodesConfigProvider
	return AuctionListSelectorArgs{
		ShardCoordinator:             argsSystemSC.ShardCoordinator,
		StakingDataProvider:          argsSystemSC.StakingDataProvider,
		MaxNodesChangeConfigProvider: nodesConfigProvider,
	}, argsSystemSC
}

func fillValidatorsInfo(t *testing.T, validatorsMap state.ShardValidatorsInfoMapHandler, sdp epochStart.StakingDataProvider) {
	for _, validator := range validatorsMap.GetAllValidatorsInfo() {
		err := sdp.FillValidatorInfo(validator.GetPublicKey())
		require.Nil(t, err)
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
	fillValidatorsInfo(t, validatorsInfo, argsSystemSC.StakingDataProvider)

	als, _ := NewAuctionListSelector(args)
	err := als.SelectNodesFromAuctionList(validatorsInfo, nil, []byte("rnd"))
	require.Nil(t, err)

	expectedValidatorsInfo := map[uint32][]state.ValidatorInfoHandler{
		0: {
			createValidatorInfo(owner1StakedKeys[0], common.EligibleList, owner1, 0),
			createValidatorInfo(owner2StakedKeys[0], common.AuctionList, owner2, 0),
		},
	}
	require.Equal(t, expectedValidatorsInfo, validatorsInfo.GetShardValidatorsInfoMap())
}

func TestAuctionListSelector_calcSoftAuctionNodesConfig(t *testing.T) {
	t.Parallel()

	v1 := &state.ValidatorInfo{PublicKey: []byte("pk1")}
	v2 := &state.ValidatorInfo{PublicKey: []byte("pk2")}
	v3 := &state.ValidatorInfo{PublicKey: []byte("pk3")}
	v4 := &state.ValidatorInfo{PublicKey: []byte("pk4")}
	v5 := &state.ValidatorInfo{PublicKey: []byte("pk5")}
	v6 := &state.ValidatorInfo{PublicKey: []byte("pk6")}
	v7 := &state.ValidatorInfo{PublicKey: []byte("pk7")}
	v8 := &state.ValidatorInfo{PublicKey: []byte("pk8")}

	ownersData := map[string]*ownerData{
		"owner1": {
			numActiveNodes:           2,
			numAuctionNodes:          2,
			numQualifiedAuctionNodes: 2,
			numStakedNodes:           4,
			totalTopUp:               big.NewInt(1500),
			topUpPerNode:             big.NewInt(375),
			qualifiedTopUpPerNode:    big.NewInt(375),
			auctionList:              []state.ValidatorInfoHandler{v1, v2},
		},
		"owner2": {
			numActiveNodes:           0,
			numAuctionNodes:          3,
			numQualifiedAuctionNodes: 3,
			numStakedNodes:           3,
			totalTopUp:               big.NewInt(3000),
			topUpPerNode:             big.NewInt(1000),
			qualifiedTopUpPerNode:    big.NewInt(1000),
			auctionList:              []state.ValidatorInfoHandler{v3, v4, v5},
		},
		"owner3": {
			numActiveNodes:           1,
			numAuctionNodes:          2,
			numQualifiedAuctionNodes: 2,
			numStakedNodes:           3,
			totalTopUp:               big.NewInt(1000),
			topUpPerNode:             big.NewInt(333),
			qualifiedTopUpPerNode:    big.NewInt(333),
			auctionList:              []state.ValidatorInfoHandler{v6, v7},
		},
		"owner4": {
			numActiveNodes:           1,
			numAuctionNodes:          1,
			numQualifiedAuctionNodes: 1,
			numStakedNodes:           2,
			totalTopUp:               big.NewInt(0),
			topUpPerNode:             big.NewInt(0),
			qualifiedTopUpPerNode:    big.NewInt(0),
			auctionList:              []state.ValidatorInfoHandler{v8},
		},
	}

	minTopUp, maxTopUp := getMinMaxPossibleTopUp(ownersData)
	require.Equal(t, big.NewInt(1), minTopUp)    // owner3 having all nodes in auction
	require.Equal(t, big.NewInt(3000), maxTopUp) // owner2 having only only one node in auction

	softAuctionConfig, err := calcSoftAuctionNodesConfig(ownersData, 10)
	require.Nil(t, err)
	require.Equal(t, ownersData, softAuctionConfig) // 7 nodes in auction and 10 available slots; everyone gets selected

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 9)
	require.Nil(t, err)
	require.Equal(t, ownersData, softAuctionConfig) // 7 nodes in auction and 10 available slots; everyone gets selected

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 8)
	displayOwnersSelectedNodes(softAuctionConfig)
	require.Nil(t, err)
	require.Equal(t, ownersData, softAuctionConfig) // 7 nodes in auction and 8 available slots; everyone gets selected

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 7)
	expectedConfig := copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig) // 7 nodes in auction and 7 available slots; everyone gets selected

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 6)
	displayOwnersSelectedNodes(softAuctionConfig)
	expectedConfig = copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	expectedConfig["owner3"].numQualifiedAuctionNodes = 1
	expectedConfig["owner3"].qualifiedTopUpPerNode = big.NewInt(500)
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig)

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 5)
	displayOwnersSelectedNodes(softAuctionConfig)
	expectedConfig = copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	expectedConfig["owner3"].numQualifiedAuctionNodes = 1
	expectedConfig["owner3"].qualifiedTopUpPerNode = big.NewInt(500)
	expectedConfig["owner1"].numQualifiedAuctionNodes = 1
	expectedConfig["owner1"].qualifiedTopUpPerNode = big.NewInt(500)
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig)

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 4)
	displayOwnersSelectedNodes(softAuctionConfig)
	expectedConfig = copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	expectedConfig["owner3"].numQualifiedAuctionNodes = 1
	expectedConfig["owner3"].qualifiedTopUpPerNode = big.NewInt(500)
	expectedConfig["owner1"].numQualifiedAuctionNodes = 1
	expectedConfig["owner1"].qualifiedTopUpPerNode = big.NewInt(500)
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig)

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 3)
	displayOwnersSelectedNodes(softAuctionConfig)
	expectedConfig = copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	delete(expectedConfig, "owner1")
	delete(expectedConfig, "owner3")
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig)

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 2)
	displayOwnersSelectedNodes(softAuctionConfig)
	expectedConfig = copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	delete(expectedConfig, "owner1")
	delete(expectedConfig, "owner3")
	expectedConfig["owner2"].numQualifiedAuctionNodes = 2
	expectedConfig["owner2"].qualifiedTopUpPerNode = big.NewInt(1500)
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig)

	softAuctionConfig, err = calcSoftAuctionNodesConfig(ownersData, 1)
	displayOwnersSelectedNodes(softAuctionConfig)
	expectedConfig = copyOwnersData(ownersData)
	delete(expectedConfig, "owner4")
	delete(expectedConfig, "owner1")
	delete(expectedConfig, "owner3")
	expectedConfig["owner2"].numQualifiedAuctionNodes = 1
	expectedConfig["owner2"].qualifiedTopUpPerNode = big.NewInt(3000)
	require.Nil(t, err)
	require.Equal(t, expectedConfig, softAuctionConfig)
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

func TestCalcNormalizedRandomness(t *testing.T) {
	t.Parallel()

	t.Run("randomness longer than expected len", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 2)
		require.Equal(t, []byte("ra"), result)
	})

	t.Run("randomness length equal to expected len", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 4)
		require.Equal(t, []byte("rand"), result)
	})

	t.Run("randomness length less than expected len", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 6)
		require.Equal(t, []byte("randra"), result)
	})

	t.Run("expected len is zero", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 0)
		require.Empty(t, result)
	})
}
