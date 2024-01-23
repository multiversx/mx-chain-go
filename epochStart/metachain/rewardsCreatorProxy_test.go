package metachain

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/require"
)

func TestNewRewardsCreatorProxy_MissingRewardsCreatorV1ArgumentShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultRewardsCreatorProxyArgs()
	args.ShardCoordinator = nil
	rewardsCreatorProxy, err := NewRewardsCreatorProxy(args)
	require.Equal(t, epochStart.ErrNilShardCoordinator, err)
	require.Nil(t, rewardsCreatorProxy)
}

func TestNewRewardsCreatorProxy_MissingRewardsCreatorV2ArgumentShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultRewardsCreatorProxyArgs()
	args.StakingDataProvider = nil
	rewardsCreatorProxy, err := NewRewardsCreatorProxy(args)
	require.Equal(t, epochStart.ErrNilStakingDataProvider, err)
	require.Nil(t, rewardsCreatorProxy)
}

func TestNewRewardsCreatorProxy_OK(t *testing.T) {
	t.Parallel()

	args := createDefaultRewardsCreatorProxyArgs()
	rewardsCreatorProxy, err := NewRewardsCreatorProxy(args)
	require.Nil(t, err)
	require.NotNil(t, rewardsCreatorProxy)
	require.Equal(t, rCreatorV1, rewardsCreatorProxy.configuredRC)
}

func TestRewardsCreatorProxy_CreateRewardsMiniBlocksWithError(t *testing.T) {
	t.Parallel()
	expectedErr := fmt.Errorf("expectedError")

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		CreateRewardsMiniBlocksCalled: func(
			metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
		) (block.MiniBlockSlice, error) {
			return nil, expectedErr
		},
	}

	rewardsCreatorProxy, vInfo, mb := createTestData(rewardCreatorV1, rCreatorV1)
	computedEconomics := &mb.EpochStart.Economics

	miniBlocks, err := rewardsCreatorProxy.CreateRewardsMiniBlocks(mb, vInfo, computedEconomics)
	require.Equal(t, expectedErr, err)
	require.Nil(t, miniBlocks)
}

func TestRewardsCreatorProxy_CreateRewardsMiniBlocksOK(t *testing.T) {
	t.Parallel()

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		CreateRewardsMiniBlocksCalled: func(
			metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
		) (block.MiniBlockSlice, error) {
			return make(block.MiniBlockSlice, 2), nil
		},
	}

	rewardsCreatorProxy, vInfo, mb := createTestData(rewardCreatorV1, rCreatorV1)
	economics := &mb.EpochStart.Economics

	miniBlocks, err := rewardsCreatorProxy.CreateRewardsMiniBlocks(mb, vInfo, economics)
	require.Nil(t, err)
	require.NotNil(t, miniBlocks)
}

func TestRewardsCreatorProxy_CreateRewardsMiniBlocksWithSwitchToRewardsCreatorV2(t *testing.T) {
	t.Parallel()

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		CreateRewardsMiniBlocksCalled: func(
			metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
		) (block.MiniBlockSlice, error) {
			return make(block.MiniBlockSlice, 2), nil
		},
	}

	rewardsCreatorProxy, vInfo, metaBlock := createTestData(rewardCreatorV1, rCreatorV1)
	stub, _ := rewardsCreatorProxy.args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stub.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
		if flag == common.StakingV2FlagAfterEpoch {
			return epoch > 1
		}
		return false
	}
	metaBlock.Epoch = 3
	economics := &metaBlock.EpochStart.Economics

	miniBlocks, err := rewardsCreatorProxy.CreateRewardsMiniBlocks(metaBlock, vInfo, economics)
	require.Nil(t, err)
	require.NotNil(t, miniBlocks)
	require.Equal(t, rewardsCreatorProxy.configuredRC, rCreatorV2)
	_, ok := rewardsCreatorProxy.rc.(*rewardsCreatorV2)
	require.True(t, ok)
	_, ok = rewardsCreatorProxy.rc.(*rewardsCreator)
	require.False(t, ok)
}

func TestRewardsCreatorProxy_CreateRewardsMiniBlocksWithSwitchToRewardsCreatorV1(t *testing.T) {
	t.Parallel()

	rewardCreatorV2 := &mock.RewardsCreatorStub{
		CreateRewardsMiniBlocksCalled: func(
			metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
		) (block.MiniBlockSlice, error) {
			return make(block.MiniBlockSlice, 2), nil
		},
	}

	rewardsCreatorProxy, vInfo, metaBlock := createTestData(rewardCreatorV2, rCreatorV2)
	stub, _ := rewardsCreatorProxy.args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stub.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
		if flag == common.StakingV2FlagAfterEpoch {
			return epoch > 5
		}
		return false
	}

	metaBlock.Epoch = 3
	economics := &metaBlock.EpochStart.Economics

	miniBlocks, err := rewardsCreatorProxy.CreateRewardsMiniBlocks(metaBlock, vInfo, economics)
	require.Nil(t, err)
	require.NotNil(t, miniBlocks)
	require.Equal(t, rewardsCreatorProxy.configuredRC, rCreatorV1)
	_, ok := rewardsCreatorProxy.rc.(*rewardsCreator)
	require.True(t, ok)
	_, ok = rewardsCreatorProxy.rc.(*rewardsCreatorV2)
	require.False(t, ok)
}

func TestRewardsCreatorProxy_VerifyRewardsMiniBlocksWithError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expectedError")
	rewardCreatorV1 := &mock.RewardsCreatorStub{
		VerifyRewardsMiniBlocksCalled: func(
			metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics) error {
			return expectedErr
		},
	}

	rewardsCreatorProxy, vInfo, mb := createTestData(rewardCreatorV1, rCreatorV1)
	economics := &mb.EpochStart.Economics

	err := rewardsCreatorProxy.VerifyRewardsMiniBlocks(mb, vInfo, economics)
	require.Equal(t, expectedErr, err)
}

func TestRewardsCreatorProxy_VerifyRewardsMiniBlocksOK(t *testing.T) {
	t.Parallel()

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		VerifyRewardsMiniBlocksCalled: func(
			metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics) error {
			return nil
		},
	}

	rewardsCreatorProxy, vInfo, mb := createTestData(rewardCreatorV1, rCreatorV1)
	economics := &mb.EpochStart.Economics

	err := rewardsCreatorProxy.VerifyRewardsMiniBlocks(mb, vInfo, economics)
	require.Nil(t, err)
}

func TestRewardsCreatorProxy_GetProtocolSustainabilityRewards(t *testing.T) {
	t.Parallel()

	expectedValue := big.NewInt(12345)
	rewardCreatorV1 := &mock.RewardsCreatorStub{
		GetProtocolSustainabilityRewardsCalled: func() *big.Int {
			return expectedValue
		},
	}

	rewardsCreatorProxy, _, _ := createTestData(rewardCreatorV1, rCreatorV1)

	protocolSustainabilityRewards := rewardsCreatorProxy.GetProtocolSustainabilityRewards()
	require.Equal(t, expectedValue, protocolSustainabilityRewards)
}

func TestRewardsCreatorProxy_GetLocalTxCache(t *testing.T) {
	t.Parallel()

	expectedValue := &mock.TxForCurrentBlockStub{}
	rewardCreatorV1 := &mock.RewardsCreatorStub{
		GetLocalTxCacheCalled: func() epochStart.TransactionCacher {
			return expectedValue
		},
	}

	rewardsCreatorProxy, _, _ := createTestData(rewardCreatorV1, rCreatorV1)

	protocolSustainabilityRewards := rewardsCreatorProxy.GetLocalTxCache()
	require.Equal(t, expectedValue, protocolSustainabilityRewards)
}

func TestRewardsCreatorProxy_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	expectedValue := make(map[string][][]byte)
	blockBody := createDefaultBlockBody()

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		CreateMarshalledDataCalled: func(body *block.Body) map[string][][]byte {
			if blockBody == body {
				return expectedValue
			}
			return nil
		},
	}

	rewardsCreatorProxy, _, _ := createTestData(rewardCreatorV1, rCreatorV1)

	protocolSustainabilityRewards := rewardsCreatorProxy.CreateMarshalledData(blockBody)
	require.Equal(t, expectedValue, protocolSustainabilityRewards)
}

func TestRewardsCreatorProxy_GetRewardsTxs(t *testing.T) {
	t.Parallel()

	expectedValue := make(map[string]data.TransactionHandler)
	expectedValue["testkey"] = &rewardTx.RewardTx{
		Value: big.NewInt(100),
	}
	blockBody := createDefaultBlockBody()

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		GetRewardsTxsCalled: func(body *block.Body) map[string]data.TransactionHandler {
			if blockBody == body {
				return expectedValue
			}
			return nil
		},
	}

	rewardsCreatorProxy, _, _ := createTestData(rewardCreatorV1, rCreatorV1)

	protocolSustainabilityRewards := rewardsCreatorProxy.GetRewardsTxs(blockBody)
	require.Equal(t, expectedValue, protocolSustainabilityRewards)
}

func TestRewardsCreatorProxy_SaveTxBlockToStorage(t *testing.T) {
	t.Parallel()

	blockBody := createDefaultBlockBody()
	functionCalled := false

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		SaveBlockDataToStorageCalled: func(metaBlock data.MetaHeaderHandler, body *block.Body) {
			functionCalled = true
		},
	}

	rewardsCreatorProxy, _, metaBlock := createTestData(rewardCreatorV1, rCreatorV1)

	rewardsCreatorProxy.SaveBlockDataToStorage(metaBlock, blockBody)
	require.Equal(t, true, functionCalled)
}

func TestRewardsCreatorProxy_DeleteTxsFromStorage(t *testing.T) {
	t.Parallel()

	blockBody := createDefaultBlockBody()
	functionCalled := false

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		DeleteBlockDataFromStorageCalled: func(metaBlock data.MetaHeaderHandler, body *block.Body) {
			functionCalled = true
		},
	}

	rewardsCreatorProxy, _, metaBlock := createTestData(rewardCreatorV1, rCreatorV1)

	rewardsCreatorProxy.DeleteBlockDataFromStorage(metaBlock, blockBody)
	require.Equal(t, true, functionCalled)
}

func TestRewardsCreatorProxy_RemoveBlockDataFromPools(t *testing.T) {
	t.Parallel()

	blockBody := createDefaultBlockBody()
	functionCalled := false

	rewardCreatorV1 := &mock.RewardsCreatorStub{
		RemoveBlockDataFromPoolsCalled: func(metaBlock data.MetaHeaderHandler, body *block.Body) {
			functionCalled = true
		},
	}

	rewardsCreatorProxy, _, metaBlock := createTestData(rewardCreatorV1, rCreatorV1)

	rewardsCreatorProxy.RemoveBlockDataFromPools(metaBlock, blockBody)
	require.Equal(t, true, functionCalled)
}

func TestRewardsCreatorProxy_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var rewardsCreatorProxy epochStart.RewardsCreator
	require.True(t, check.IfNil(rewardsCreatorProxy))

	rewardCreatorV1 := &mock.RewardsCreatorStub{}
	rewardsCreatorProxy, _, _ = createTestData(rewardCreatorV1, rCreatorV1)

	require.False(t, check.IfNil(rewardsCreatorProxy))
}

func createTestData(rewardCreator *mock.RewardsCreatorStub, rcType configuredRewardsCreator) (*rewardsCreatorProxy, map[uint32][]*state.ValidatorInfo, *block.MetaBlock) {
	args := createDefaultRewardsCreatorProxyArgs()
	rewardsCreatorProxy := &rewardsCreatorProxy{
		rc:           rewardCreator,
		configuredRC: rcType,
		args:         &args,
	}

	vInfo := createDefaultValidatorInfo(400, args.ShardCoordinator, args.NodesConfigProvider, 100, uint32(14400))
	mb := createDummyMetaBlock()
	return rewardsCreatorProxy, vInfo, mb
}

func createDefaultBlockBody() *block.Body {
	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{},
		Epoch:   0,
	}
	rwdTxHash, _ := core.CalculateHash(&marshal.JsonMarshalizer{}, &hashingMocks.HasherMock{}, rwdTx)

	return &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				ReceiverShardID: 0,
				Type:            block.RewardsBlock,
				TxHashes:        [][]byte{rwdTxHash},
			},
		},
	}
}

func createDefaultRewardsCreatorProxyArgs() RewardsCreatorProxyArgs {
	rewardsTopUpGradientPoint, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)

	rewardsHandler := &economicsmocks.EconomicsHandlerStub{
		RewardsTopUpGradientPointCalled: func() *big.Int {
			return rewardsTopUpGradientPoint
		},
		RewardsTopUpFactorCalled: func() float64 {
			return 0.25
		},
	}

	return RewardsCreatorProxyArgs{
		BaseRewardsCreatorArgs: getBaseRewardsArguments(),
		StakingDataProvider:    &mock.StakingDataProviderStub{},
		EconomicsDataProvider:  NewEpochEconomicsStatistics(),
		RewardsHandler:         rewardsHandler,
	}
}

func createDummyMetaBlock() *block.MetaBlock {
	return &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
}
