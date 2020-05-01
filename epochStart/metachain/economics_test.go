package metachain

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockEpochEconomicsArguments() ArgsNewEpochEconomics {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)

	argsNewEpochEconomics := ArgsNewEpochEconomics{
		Hasher:              &mock.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		Store:               createMetaStore(),
		ShardCoordinator:    shardCoordinator,
		NodesConfigProvider: &mock.NodesCoordinatorStub{},
		RewardsHandler:      &mock.RewardsHandlerStub{},
		RoundTime:           &mock.RoundTimeDurationHandler{},
	}
	return argsNewEpochEconomics
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.Marshalizer = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilStore(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.Store = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, epochStart.ErrNilStorage, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.ShardCoordinator = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, epochStart.ErrNilShardCoordinator, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilNodesdCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.NodesConfigProvider = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, epochStart.ErrNilNodesConfigProvider, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilRewardsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.RewardsHandler = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, epochStart.ErrNilRewardsHandler, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorNilRounder(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()
	arguments.RoundTime = nil

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.Nil(t, esd)
	require.Equal(t, epochStart.ErrNilRounder, err)
}

func TestEpochEconomics_NewEndOfEpochEconomicsDataCreatorShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochEconomicsArguments()

	esd, err := NewEndOfEpochEconomicsDataCreator(arguments)
	require.NotNil(t, esd)
	require.Nil(t, err)
}

func TestNewEndOfEpochEconomicsDataCreator_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Marshalizer = nil
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.True(t, check.IfNil(eoeedc))
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestNewEndOfEpochEconomicsDataCreator_NilStore(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Store = nil
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.True(t, check.IfNil(eoeedc))
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestNewEndOfEpochEconomicsDataCreator_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.ShardCoordinator = nil
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.True(t, check.IfNil(eoeedc))
	assert.Equal(t, epochStart.ErrNilShardCoordinator, err)
}

func TestNewEndOfEpochEconomicsDataCreator_NilNodesCoordinator(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.NodesConfigProvider = nil
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.True(t, check.IfNil(eoeedc))
	assert.Equal(t, epochStart.ErrNilNodesConfigProvider, err)
}

func TestNewEndOfEpochEconomicsDataCreator_NilRewardsHandler(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.RewardsHandler = nil
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.True(t, check.IfNil(eoeedc))
	assert.Equal(t, epochStart.ErrNilRewardsHandler, err)
}

func TestNewEndOfEpochEconomicsDataCreator_NilRoundTimeDurationHandler(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.RoundTime = nil
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.True(t, check.IfNil(eoeedc))
	assert.Equal(t, epochStart.ErrNilRounder, err)
}

func TestNewEndOfEpochEconomicsDataCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArguments()
	eoeedc, err := NewEndOfEpochEconomicsDataCreator(args)

	assert.False(t, check.IfNil(eoeedc))
	assert.Nil(t, err)
}

func TestEconomics_ComputeEndOfEpochEconomics_NilMetaBlockShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	res, err := ec.ComputeEndOfEpochEconomics(nil)
	assert.Nil(t, res)
	assert.Equal(t, epochStart.ErrNilHeaderHandler, err)
}

func TestEconomics_ComputeEndOfEpochEconomics_NilAccumulatedFeesInEpochShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	mb := block.MetaBlock{AccumulatedFeesInEpoch: nil}
	res, err := ec.ComputeEndOfEpochEconomics(&mb)
	assert.Nil(t, res)
	assert.Equal(t, epochStart.ErrNilTotalAccumulatedFeesInEpoch, err)
}

func TestEconomics_ComputeRewardsForCommunity(t *testing.T) {
	totalRewards := big.NewInt(0).SetUint64(123456)
	args := getArguments()
	args.RewardsHandler = &mock.RewardsHandlerStub{
		CommunityPercentageCalled: func() float64 {
			return 0.1
		},
	}
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	expectedRewards := big.NewInt(12345)
	communityRewards := ec.computeRewardsForCommunity(totalRewards)
	assert.Equal(t, expectedRewards, communityRewards)
}

func TestEconomics_AdjustRewardsPerBlockWithCommunityRewards(t *testing.T) {
	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rwdPerBlock := big.NewInt(0).SetUint64(1000)
	blocksInEpoch := uint64(100)
	communityRewards := big.NewInt(0).SetUint64(10000)

	expectedRewardsCommunityAfterAdjustment := big.NewInt(0).Set(communityRewards)
	expectedRwdPerBlock := big.NewInt(900)

	ec.adjustRewardsPerBlockWithCommunityRewards(rwdPerBlock, communityRewards, blocksInEpoch)

	assert.Equal(t, expectedRewardsCommunityAfterAdjustment, communityRewards)
	assert.Equal(t, expectedRwdPerBlock, rwdPerBlock)
}

func TestEconomics_AdjustRewardsPerBlockWithLeaderPercentage(t *testing.T) {
	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rwdPerBlock := big.NewInt(0).SetUint64(1000)
	blocksInEpoch := uint64(100)
	accumulatedFeesInEpoch := big.NewInt(0).SetUint64(10000)

	expectedRwdPerBlock := big.NewInt(900)
	ec.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, accumulatedFeesInEpoch, blocksInEpoch)
	assert.Equal(t, expectedRwdPerBlock, rwdPerBlock)
}

func TestEconomics_ComputeEndOfEpochEconomics_NotEpochStartShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	mb1 := block.MetaBlock{
		Epoch:                  0,
		AccumulatedFeesInEpoch: big.NewInt(10),
	}
	res, err := ec.ComputeEndOfEpochEconomics(&mb1)
	assert.Nil(t, res)
	assert.Equal(t, epochStart.ErrNotEpochStartBlock, err)
}

func TestEconomics_ComputeEndOfEpochEconomics(t *testing.T) {
	t.Parallel()

	args := getArguments()
	args.Store = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{GetCalled: func(key []byte) ([]byte, error) {
				hdr := block.MetaBlock{
					Round: 10,
					Nonce: 5,
					EpochStart: block.EpochStart{
						Economics: block.Economics{
							TotalSupply:            big.NewInt(100000),
							TotalToDistribute:      big.NewInt(10),
							TotalNewlyMinted:       big.NewInt(109),
							RewardsPerBlockPerNode: big.NewInt(10),
							NodePrice:              big.NewInt(10),
						},
					},
				}
				hdrBytes, _ := json.Marshal(hdr)
				return hdrBytes, nil
			}}
		},
	}
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	mb := block.MetaBlock{
		Round: 15000,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, Round: 2, Nonce: 3},
				{ShardID: 1, Round: 2, Nonce: 3},
			},
			Economics: block.Economics{},
		},
		Epoch:                  2,
		AccumulatedFeesInEpoch: big.NewInt(10000),
	}
	res, err := ec.ComputeEndOfEpochEconomics(&mb)
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestEconomics_VerifyRewardsPerBlock_DifferentHitRates(t *testing.T) {
	t.Parallel()

	commAddress := "communityAddress"
	totalSupply := big.NewInt(20000000000) // 20B
	accFeesInEpoch := big.NewInt(0)
	roundDur := 4
	args := getArguments()
	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func() float64 {
			return 0.1
		},
		CommunityAddressCalled: func() string {
			return commAddress
		},
		CommunityPercentageCalled: func() float64 {
			return 0.1
		},
	}
	args.RoundTime = &mock.RoundTimeDurationHandler{
		TimeDurationCalled: func() time.Duration {
			return time.Duration(roundDur) * time.Second
		},
	}
	hdrPrevEpochStart := block.MetaBlock{
		Round: 0,
		Nonce: 0,
		Epoch: 0,
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:            totalSupply,
				TotalToDistribute:      big.NewInt(10),
				TotalNewlyMinted:       big.NewInt(10),
				RewardsPerBlockPerNode: big.NewInt(10),
				NodePrice:              big.NewInt(10),
				RewardsForCommunity:    big.NewInt(10),
			},
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, Nonce: 0},
				{ShardID: 1, Nonce: 0},
			},
		},
	}
	args.Store = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			// this will be the previous epoch meta block. It has initial 0 values so it can be considered at genesis
			return &mock.StorerStub{GetCalled: func(key []byte) ([]byte, error) {
				hdrBytes, _ := json.Marshal(hdrPrevEpochStart)
				return hdrBytes, nil
			}}
		},
	}
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	expRwdPerBlock := 84 // based on 0.1 inflation

	numBlocksInEpochSlice := []int{
		numberOfSecondsInDay / roundDur,       // 100 % hit rate
		(numberOfSecondsInDay / roundDur) / 2, // 50 % hit rate
		(numberOfSecondsInDay / roundDur) / 4, // 25 % hit rate
		(numberOfSecondsInDay / roundDur) / 8, // 12.5 % hit rate
		1,                                     // only the metablock was committed in that epoch
		37,                                    // random
		63,                                    // random
	}

	hdrPrevEpochStartHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, hdrPrevEpochStart)
	for _, numBlocksInEpoch := range numBlocksInEpochSlice {
		expectedTotalToDistribute := big.NewInt(int64(expRwdPerBlock * numBlocksInEpoch * 3)) // 2 shards + meta
		expectedTotalNewlyMinted := big.NewInt(0).Sub(expectedTotalToDistribute, accFeesInEpoch)
		expectedTotalSupply := big.NewInt(0).Add(totalSupply, expectedTotalNewlyMinted)
		expectedCommunityRewards := big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(10))
		commRewardPerBlock := big.NewInt(0).Div(expectedCommunityRewards, big.NewInt(int64(numBlocksInEpoch*3)))
		adjustedRwdPerBlock := big.NewInt(0).Sub(big.NewInt(int64(expRwdPerBlock)), commRewardPerBlock)

		mb := block.MetaBlock{
			Round: uint64(numBlocksInEpoch),
			Nonce: uint64(numBlocksInEpoch),
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0, Round: uint64(numBlocksInEpoch), Nonce: uint64(numBlocksInEpoch)},
					{ShardID: 1, Round: uint64(numBlocksInEpoch), Nonce: uint64(numBlocksInEpoch)},
				},
				Economics: block.Economics{
					TotalSupply:            expectedTotalSupply,
					TotalToDistribute:      expectedTotalToDistribute,
					TotalNewlyMinted:       expectedTotalNewlyMinted,
					RewardsPerBlockPerNode: adjustedRwdPerBlock,
					NodePrice:              big.NewInt(10),
					PrevEpochStartHash:     hdrPrevEpochStartHash,
					RewardsForCommunity:    expectedCommunityRewards,
				},
			},
			Epoch:                  1,
			AccumulatedFeesInEpoch: accFeesInEpoch,
		}

		err := ec.VerifyRewardsPerBlock(&mb)
		assert.Nil(t, err)
	}
}

func getArguments() ArgsNewEpochEconomics {
	return ArgsNewEpochEconomics{
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              mock.HasherMock{},
		Store:               &mock.ChainStorerStub{},
		ShardCoordinator:    mock.NewMultipleShardsCoordinatorMock(),
		NodesConfigProvider: &mock.NodesCoordinatorStub{},
		RewardsHandler:      &mock.RewardsHandlerStub{},
		RoundTime:           &mock.RoundTimeDurationHandler{},
	}
}
