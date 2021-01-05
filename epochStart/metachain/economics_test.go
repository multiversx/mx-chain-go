package metachain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockEpochEconomicsArguments() ArgsNewEpochEconomics {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)

	argsNewEpochEconomics := ArgsNewEpochEconomics{
		Hasher:                &mock.HasherMock{},
		Marshalizer:           &mock.MarshalizerMock{},
		Store:                 createMetaStore(),
		ShardCoordinator:      shardCoordinator,
		RewardsHandler:        &mock.RewardsHandlerStub{},
		RoundTime:             &mock.RoundTimeDurationHandler{},
		GenesisTotalSupply:    big.NewInt(2000000),
		EconomicsDataNotified: NewEpochEconomicsStatistics(),
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

	mb := block.MetaBlock{AccumulatedFeesInEpoch: nil, DevFeesInEpoch: big.NewInt(0)}
	res, err := ec.ComputeEndOfEpochEconomics(&mb)
	assert.Nil(t, res)
	assert.Equal(t, epochStart.ErrNilTotalAccumulatedFeesInEpoch, err)
}

func TestEconomics_ComputeEndOfEpochEconomics_NilDevFeesInEpochShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	mb := block.MetaBlock{AccumulatedFeesInEpoch: big.NewInt(0), DevFeesInEpoch: nil}
	res, err := ec.ComputeEndOfEpochEconomics(&mb)
	assert.Nil(t, res)
	assert.Equal(t, epochStart.ErrNilTotalDevFeesInEpoch, err)
}

func TestEconomics_ComputeRewardsForProtocolSustainability(t *testing.T) {
	totalRewards := big.NewInt(0).SetUint64(123456)
	args := getArguments()
	args.RewardsHandler = &mock.RewardsHandlerStub{
		ProtocolSustainabilityPercentageCalled: func() float64 {
			return 0.1
		},
	}
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	expectedRewards := big.NewInt(12345)
	protocolSustainabilityRewards := ec.computeRewardsForProtocolSustainability(totalRewards)
	assert.Equal(t, expectedRewards, protocolSustainabilityRewards)
}

func TestEconomics_AdjustRewardsPerBlockWithProtocolSustainabilityRewards(t *testing.T) {
	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rwdPerBlock := big.NewInt(0).SetUint64(1000)
	blocksInEpoch := uint64(100)
	protocolSustainabilityRewards := big.NewInt(0).SetUint64(10000)

	expectedRewardsProtocolSustainabilityAfterAdjustment := big.NewInt(0).Set(protocolSustainabilityRewards)
	expectedRwdPerBlock := big.NewInt(900)

	ec.adjustRewardsPerBlockWithProtocolSustainabilityRewards(rwdPerBlock, protocolSustainabilityRewards, blocksInEpoch)

	assert.Equal(t, expectedRewardsProtocolSustainabilityAfterAdjustment, protocolSustainabilityRewards)
	assert.Equal(t, expectedRwdPerBlock, rwdPerBlock)
}

func TestEconomics_AdjustRewardsPerBlockWithLeaderPercentage(t *testing.T) {
	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rwdPerBlock := big.NewInt(0).SetUint64(1000)
	blocksInEpoch := uint64(100)
	accumulatedFeesInEpoch := big.NewInt(0).SetUint64(10000)

	expectedRwdPerBlock := big.NewInt(900)
	ec.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, accumulatedFeesInEpoch, big.NewInt(0), blocksInEpoch, 1)
	assert.Equal(t, expectedRwdPerBlock, rwdPerBlock)
}

func TestEconomics_AdjustRewardsPerBlockWithLeaderPercentageAndDevFees(t *testing.T) {
	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rwdPerBlock := big.NewInt(0).SetUint64(1000)
	blocksInEpoch := uint64(100)
	accumulatedFeesInEpoch := big.NewInt(0).SetUint64(10000)
	devFeesInEpoch := big.NewInt(1000)

	expectedRwdPerBlock := big.NewInt(910)
	ec.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, accumulatedFeesInEpoch, devFeesInEpoch, blocksInEpoch, 1)
	assert.Equal(t, expectedRwdPerBlock, rwdPerBlock)
}

func TestEconomics_AdjustRewardsPerBlockWithDeveloperFees(t *testing.T) {
	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rwdPerBlock := big.NewInt(0).SetUint64(1000)
	blocksInEpoch := uint64(100)
	developerFees := big.NewInt(0).SetUint64(500)

	expectedRwdPerBlock := big.NewInt(995)
	ec.adjustRewardsPerBlockWithDeveloperFees(rwdPerBlock, developerFees, blocksInEpoch)
	assert.Equal(t, expectedRwdPerBlock, rwdPerBlock)
}

func TestEconomics_ComputeEndOfEpochEconomics_NotEpochStartShouldErr(t *testing.T) {
	t.Parallel()

	args := getArguments()
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	mb1 := block.MetaBlock{
		Epoch:                  0,
		AccumulatedFeesInEpoch: big.NewInt(10),
		DevFeesInEpoch:         big.NewInt(0),
	}
	res, err := ec.ComputeEndOfEpochEconomics(&mb1)
	assert.Nil(t, res)
	assert.Equal(t, epochStart.ErrNotEpochStartBlock, err)
}

func TestEconomics_ComputeInflationRate(t *testing.T) {
	args := getArguments()
	errNotGoodYear := errors.New("not good year")
	var errFound error
	year1inflation := 1.0
	year2inflation := 0.5
	lateYearInflation := 2.0

	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func(year uint32) float64 {
			switch year {
			case 0:
				errFound = errNotGoodYear
				return 0.0
			case 1:
				return year1inflation
			case 2:
				return year2inflation
			default:
				return lateYearInflation
			}
		},
	}
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	rate := ec.computeInflationRate(1)
	assert.Nil(t, errFound)
	assert.Equal(t, rate, year1inflation)

	rate = ec.computeInflationRate(50000)
	assert.Nil(t, errFound)
	assert.Equal(t, rate, year1inflation)

	rate = ec.computeInflationRate(7884000)
	assert.Nil(t, errFound)
	assert.Equal(t, rate, year2inflation)

	rate = ec.computeInflationRate(8884000)
	assert.Nil(t, errFound)
	assert.Equal(t, rate, year2inflation)

	rate = ec.computeInflationRate(38884000)
	assert.Nil(t, errFound)
	assert.Equal(t, rate, lateYearInflation)
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
							TotalSupply:       big.NewInt(100000),
							TotalToDistribute: big.NewInt(10),
							TotalNewlyMinted:  big.NewInt(109),
							RewardsPerBlock:   big.NewInt(10),
							NodePrice:         big.NewInt(10),
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
		DevFeesInEpoch:         big.NewInt(0),
	}
	res, err := ec.ComputeEndOfEpochEconomics(&mb)
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestEconomics_VerifyRewardsPerBlock_DifferentHitRates(t *testing.T) {
	t.Parallel()

	commAddress := "protocolSustainabilityAddress"
	totalSupply := big.NewInt(20000000000) // 20B
	accFeesInEpoch := big.NewInt(0)
	devFeesInEpoch := big.NewInt(0)
	roundDur := 4
	args := getArguments()
	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func(_ uint32) float64 {
			return 0.1
		},
		ProtocolSustainabilityAddressCalled: func() string {
			return commAddress
		},
		ProtocolSustainabilityPercentageCalled: func() float64 {
			return 0.1
		},
	}
	args.RoundTime = &mock.RoundTimeDurationHandler{
		TimeDurationCalled: func() time.Duration {
			return time.Duration(roundDur) * time.Second
		},
	}
	newTotalSupply := big.NewInt(0).Add(totalSupply, totalSupply)
	hdrPrevEpochStart := block.MetaBlock{
		Round: 0,
		Nonce: 0,
		Epoch: 0,
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:                      newTotalSupply,
				TotalToDistribute:                big.NewInt(10),
				TotalNewlyMinted:                 big.NewInt(10),
				RewardsPerBlock:                  big.NewInt(10),
				NodePrice:                        big.NewInt(10),
				RewardsForProtocolSustainability: big.NewInt(10),
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
	args.GenesisTotalSupply = totalSupply
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
		expectedTotalSupply := big.NewInt(0).Add(newTotalSupply, expectedTotalNewlyMinted)
		expectedProtocolSustainabilityRewards := big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(10))
		commRewardPerBlock := big.NewInt(0).Div(expectedProtocolSustainabilityRewards, big.NewInt(int64(numBlocksInEpoch*3)))
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
					TotalSupply:                      expectedTotalSupply,
					TotalToDistribute:                expectedTotalToDistribute,
					TotalNewlyMinted:                 expectedTotalNewlyMinted,
					RewardsPerBlock:                  adjustedRwdPerBlock,
					NodePrice:                        big.NewInt(10),
					PrevEpochStartHash:               hdrPrevEpochStartHash,
					RewardsForProtocolSustainability: expectedProtocolSustainabilityRewards,
				},
			},
			Epoch:                  1,
			AccumulatedFeesInEpoch: accFeesInEpoch,
			DevFeesInEpoch:         devFeesInEpoch,
		}

		computedEconomics, err := ec.ComputeEndOfEpochEconomics(&mb)
		assert.Nil(t, err)
		assert.NotNil(t, computedEconomics)

		err = ec.VerifyRewardsPerBlock(&mb, expectedProtocolSustainabilityRewards, computedEconomics)
		assert.Nil(t, err)

		ecos, err := ec.ComputeEndOfEpochEconomics(&mb)
		assert.Nil(t, err)
		assert.True(t, expectedProtocolSustainabilityRewards.Cmp(ecos.RewardsForProtocolSustainability) == 0)
	}
}

func TestEconomics_VerifyRewardsPerBlock_DifferentFees(t *testing.T) {
	t.Parallel()

	commAddress := "protocolSustainabilityAddress"
	// based on 0.1 inflation, 14400 rounds, 3 shards, 365 epochs per year
	totalSupply := big.NewInt(100 * 4 * 14400 * 365 * 10) // 21.02B
	roundDur := 6
	args := getArguments()
	args.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func(_ uint32) float64 {
			return 0.1
		},
		ProtocolSustainabilityAddressCalled: func() string {
			return commAddress
		},
		ProtocolSustainabilityPercentageCalled: func() float64 {
			return 0.1
		},
		LeaderPercentageCalled: func() float64 {
			return 0.1
		},
	}
	args.RoundTime = &mock.RoundTimeDurationHandler{
		TimeDurationCalled: func() time.Duration {
			return time.Duration(roundDur) * time.Second
		},
	}
	newTotalSupply := big.NewInt(0).Add(totalSupply, big.NewInt(0))
	hdrPrevEpochStart := block.MetaBlock{
		Round: 0,
		Nonce: 0,
		Epoch: 0,
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:                      newTotalSupply,
				TotalToDistribute:                big.NewInt(0),
				TotalNewlyMinted:                 big.NewInt(0),
				RewardsPerBlock:                  big.NewInt(0),
				NodePrice:                        big.NewInt(10),
				RewardsForProtocolSustainability: big.NewInt(0),
			},
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, Nonce: 0},
				{ShardID: 1, Nonce: 0},
				{ShardID: 2, Nonce: 0},
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
	args.GenesisTotalSupply = totalSupply
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	numberOfShards := 4
	maxRewardPerBlock := int64(100)
	blocksPerDay := numberOfSecondsInDay / roundDur
	protocolSustainabilityRewardsForMaxBlocks := int64(numberOfSecondsInDay/roundDur) * 4 * maxRewardPerBlock / 10

	tests := []struct {
		name                                  string
		numBlocksInEpoch                      int
		accFeesInEpoch                        *big.Int
		devFeesInEpoch                        *big.Int
		expectedProtocolSustainabilityRewards *big.Int
		expectedRewardPerBlock                *big.Int
	}{
		{
			name:                                  "test1",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(0),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10),
		},
		{
			name:                                  "test2",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 1),
		},
		{
			name:                                  "test3",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 10 - 0), // 0 = 10% of (accFees - devFees),
		},
		{ // half the blocks, same reward per block but half the protocol sustainability
			name:                                  "test3",
			numBlocksInEpoch:                      blocksPerDay / 2,
			accFeesInEpoch:                        big.NewInt(0),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks / 2),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10),
		},
		{ // half the blocks, same reward per block but half the protocol sustainability and -2x10 from accFeesInEpoch
			name:                                  "test4",
			numBlocksInEpoch:                      blocksPerDay / 2,
			accFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks / 2),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 2),
		},
		{ // half the blocks, same reward per block but half the protocol sustainability and -2x10 from accFeesInEpoch. Half of accFeesInEpoch are devFees
			name:                                  "test5",
			numBlocksInEpoch:                      blocksPerDay / 2,
			accFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt((10 * 4 * int64(blocksPerDay)) / 2),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks / 2),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 10 - 1),
		},
		{ // half the blocks, same reward per block but half the protocol sustainability and -4x10 from fees
			name:                                  "test6",
			numBlocksInEpoch:                      blocksPerDay / 2,
			accFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(10 * 4 * int64(blocksPerDay)),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks / 2),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 20 - 0),
		},
		// just 1 block per epoch
		{
			name:                                  "test7",
			numBlocksInEpoch:                      1,
			accFeesInEpoch:                        big.NewInt(0),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(100 * 4 / 10),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10),
		},
		{
			name:                                  "test8",
			numBlocksInEpoch:                      1,
			accFeesInEpoch:                        big.NewInt(10 * 4),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(100 * 4 / 10),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 1),
		},
		{
			name:                                  "test9",
			numBlocksInEpoch:                      1,
			accFeesInEpoch:                        big.NewInt(10 * 4),
			devFeesInEpoch:                        big.NewInt(10 * 4),
			expectedProtocolSustainabilityRewards: big.NewInt(100 * 4 / 10),
			expectedRewardPerBlock:                big.NewInt(maxRewardPerBlock - 10 - 10 - 0),
		},
	}

	hdrPrevEpochStartHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, hdrPrevEpochStart)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalBlocksInEpoch := int64(tt.numBlocksInEpoch * numberOfShards)
			expectedTotalToDistribute := big.NewInt(maxRewardPerBlock * totalBlocksInEpoch)
			expectedTotalNewlyMinted := big.NewInt(0).Sub(expectedTotalToDistribute, tt.accFeesInEpoch)
			expectedTotalSupply := big.NewInt(0).Add(newTotalSupply, expectedTotalNewlyMinted)
			expectedProtocolSustainabilityRewards := big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(10))
			commRewardPerBlock := big.NewInt(0).Div(expectedProtocolSustainabilityRewards, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock := big.NewInt(0).Sub(big.NewInt(maxRewardPerBlock), commRewardPerBlock)
			feesForValidators := big.NewInt(0).Sub(tt.accFeesInEpoch, tt.devFeesInEpoch)
			feesForProposers := core.GetPercentageOfValue(feesForValidators, args.RewardsHandler.LeaderPercentage())
			accFeeRewardPerBlock := big.NewInt(0).Div(feesForProposers, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock = big.NewInt(0).Sub(adjustedRwdPerBlock, accFeeRewardPerBlock)
			devFeeRewardPerBlock := big.NewInt(0).Div(tt.devFeesInEpoch, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock = big.NewInt(0).Sub(adjustedRwdPerBlock, devFeeRewardPerBlock)

			mb := block.MetaBlock{
				Round: uint64(tt.numBlocksInEpoch),
				Nonce: uint64(tt.numBlocksInEpoch),
				EpochStart: block.EpochStart{
					LastFinalizedHeaders: []block.EpochStartShardData{
						{ShardID: 0, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
						{ShardID: 1, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
						{ShardID: 2, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
					},
					Economics: block.Economics{
						TotalSupply:                      expectedTotalSupply,
						TotalToDistribute:                expectedTotalToDistribute,
						TotalNewlyMinted:                 expectedTotalNewlyMinted,
						RewardsPerBlock:                  adjustedRwdPerBlock,
						NodePrice:                        big.NewInt(10),
						PrevEpochStartHash:               hdrPrevEpochStartHash,
						RewardsForProtocolSustainability: expectedProtocolSustainabilityRewards,
					},
				},
				Epoch:                  1,
				AccumulatedFeesInEpoch: tt.accFeesInEpoch,
				DevFeesInEpoch:         tt.devFeesInEpoch,
			}

			computedEconomics, err := ec.ComputeEndOfEpochEconomics(&mb)
			assert.Nil(t, err)
			assert.NotNil(t, computedEconomics)

			err = ec.VerifyRewardsPerBlock(&mb, expectedProtocolSustainabilityRewards, computedEconomics)
			assert.Nil(t, err)

			ecos, err := ec.ComputeEndOfEpochEconomics(&mb)
			assert.Nil(t, err)
			assert.True(t, expectedProtocolSustainabilityRewards.Cmp(ecos.RewardsForProtocolSustainability) == 0)

			require.Equal(t, tt.expectedProtocolSustainabilityRewards, ecos.RewardsForProtocolSustainability)
			require.Equal(t, tt.expectedRewardPerBlock, ecos.RewardsPerBlock)
		})
	}
}

func TestEconomics_VerifyRewardsPerBlock_MoreFeesThanInflation(t *testing.T) {
	t.Parallel()

	commAddress := "protocolSustainabilityAddress"
	// based on 0.1 inflation, 14400 rounds, 3 shards, 365 epochs per year
	totalSupply := big.NewInt(100 * 4 * 14400 * 365 * 10) // 21.02B
	roundDur := 6
	args := getArguments()
	args.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func(_ uint32) float64 {
			return 0.1
		},
		ProtocolSustainabilityAddressCalled: func() string {
			return commAddress
		},
		ProtocolSustainabilityPercentageCalled: func() float64 {
			return 0.1
		},
		LeaderPercentageCalled: func() float64 {
			return 0.1
		},
	}
	args.RoundTime = &mock.RoundTimeDurationHandler{
		TimeDurationCalled: func() time.Duration {
			return time.Duration(roundDur) * time.Second
		},
	}
	newTotalSupply := big.NewInt(0).Add(totalSupply, big.NewInt(0))
	hdrPrevEpochStart := block.MetaBlock{
		Round: 0,
		Nonce: 0,
		Epoch: 0,
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:                      newTotalSupply,
				TotalToDistribute:                big.NewInt(0),
				TotalNewlyMinted:                 big.NewInt(0),
				RewardsPerBlock:                  big.NewInt(0),
				NodePrice:                        big.NewInt(10),
				RewardsForProtocolSustainability: big.NewInt(0),
			},
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, Nonce: 0},
				{ShardID: 1, Nonce: 0},
				{ShardID: 2, Nonce: 0},
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
	args.GenesisTotalSupply = totalSupply
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	numberOfShards := 4
	maxBaseRewardPerBlock := int64(100)
	blocksPerDay := numberOfSecondsInDay / roundDur
	protocolSustainabilityRewardsForMaxBlocks := int64(numberOfSecondsInDay/roundDur) * 4 * maxBaseRewardPerBlock / 10
	protocolSustainabilityPerBlock := int64(10)

	tests := []struct {
		name                                  string
		numBlocksInEpoch                      int
		accFeesInEpoch                        *big.Int
		devFeesInEpoch                        *big.Int
		expectedProtocolSustainabilityRewards *big.Int
		expectedBaseRewardPerBlock            *big.Int
	}{
		{
			name:                                  "no fees",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(0),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(maxBaseRewardPerBlock - protocolSustainabilityPerBlock),
		},
		{
			name:                                  "accFees equal inflation",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(maxBaseRewardPerBlock * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(maxBaseRewardPerBlock - protocolSustainabilityPerBlock - 10),
		},
		{
			name:                                  "accFees and devFees so that rewardPerBlock is 1",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(89 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(89 * 4 * int64(blocksPerDay)),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(maxBaseRewardPerBlock - protocolSustainabilityPerBlock - 89),
		},
		{
			name:                                  "accFees and devFees so that rewardPerBlock is exactly 0",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(90 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(90 * 4 * int64(blocksPerDay)),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(maxBaseRewardPerBlock - protocolSustainabilityPerBlock - 90),
		},
		{
			name:                                  "accFees equal to inflation so baseReward should be still 0",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(100 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(30 * 4 * int64(blocksPerDay)),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(maxBaseRewardPerBlock*7/10 - protocolSustainabilityPerBlock - 7),
		},
		{
			name:                                  "accFees and devFees much larger than inflation",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(1000 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(300 * 4 * int64(blocksPerDay)),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks * 10),
			expectedBaseRewardPerBlock:            big.NewInt(maxBaseRewardPerBlock*10*7/10 - protocolSustainabilityPerBlock*10 - 70),
		},
		{
			name:                                  "200 accFees and devFees (30%) and 1 block",
			numBlocksInEpoch:                      1,
			accFeesInEpoch:                        big.NewInt(200 * 4),
			devFeesInEpoch:                        big.NewInt(60 * 4),
			expectedProtocolSustainabilityRewards: big.NewInt(80),
			expectedBaseRewardPerBlock:            big.NewInt(200*7/10 - 20 - 14),
		},
	}

	hdrPrevEpochStartHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, hdrPrevEpochStart)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalBlocksInEpoch := int64(tt.numBlocksInEpoch * numberOfShards)
			expectedTotalToDistribute := big.NewInt(maxBaseRewardPerBlock * totalBlocksInEpoch)
			if expectedTotalToDistribute.Cmp(tt.accFeesInEpoch) < 0 {
				expectedTotalToDistribute = tt.accFeesInEpoch
				maxBaseRewardPerBlock = big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(totalBlocksInEpoch)).Int64()
			}
			expectedTotalNewlyMinted := big.NewInt(0).Sub(expectedTotalToDistribute, tt.accFeesInEpoch)
			expectedTotalSupply := big.NewInt(0).Add(newTotalSupply, expectedTotalNewlyMinted)
			expectedProtocolSustainabilityRewards := big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(10))
			commRewardPerBlock := big.NewInt(0).Div(expectedProtocolSustainabilityRewards, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock := big.NewInt(0).Sub(big.NewInt(maxBaseRewardPerBlock), commRewardPerBlock)
			feesForValidators := big.NewInt(0).Sub(tt.accFeesInEpoch, tt.devFeesInEpoch)
			feesForProposers := core.GetPercentageOfValue(feesForValidators, args.RewardsHandler.LeaderPercentage())
			accFeeRewardPerBlock := big.NewInt(0).Div(feesForProposers, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock = big.NewInt(0).Sub(adjustedRwdPerBlock, accFeeRewardPerBlock)
			devFeeRewardPerBlock := big.NewInt(0).Div(tt.devFeesInEpoch, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock = big.NewInt(0).Sub(adjustedRwdPerBlock, devFeeRewardPerBlock)

			mb := block.MetaBlock{
				Round: uint64(tt.numBlocksInEpoch),
				Nonce: uint64(tt.numBlocksInEpoch),
				EpochStart: block.EpochStart{
					LastFinalizedHeaders: []block.EpochStartShardData{
						{ShardID: 0, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
						{ShardID: 1, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
						{ShardID: 2, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
					},
					Economics: block.Economics{
						TotalSupply:                      expectedTotalSupply,
						TotalToDistribute:                expectedTotalToDistribute,
						TotalNewlyMinted:                 expectedTotalNewlyMinted,
						RewardsPerBlock:                  adjustedRwdPerBlock,
						NodePrice:                        big.NewInt(10),
						PrevEpochStartHash:               hdrPrevEpochStartHash,
						RewardsForProtocolSustainability: expectedProtocolSustainabilityRewards,
					},
				},
				Epoch:                  1,
				AccumulatedFeesInEpoch: tt.accFeesInEpoch,
				DevFeesInEpoch:         tt.devFeesInEpoch,
			}

			computedEconomics, err := ec.ComputeEndOfEpochEconomics(&mb)
			assert.Nil(t, err)
			assert.NotNil(t, computedEconomics)

			err = ec.VerifyRewardsPerBlock(&mb, expectedProtocolSustainabilityRewards, computedEconomics)
			assert.Nil(t, err)

			ecos, err := ec.ComputeEndOfEpochEconomics(&mb)
			assert.Nil(t, err)
			assert.True(t, expectedProtocolSustainabilityRewards.Cmp(ecos.RewardsForProtocolSustainability) == 0)

			require.Equal(t, tt.expectedProtocolSustainabilityRewards, ecos.RewardsForProtocolSustainability)
			require.Equal(t, tt.expectedBaseRewardPerBlock.Int64(), ecos.RewardsPerBlock.Int64())
		})
	}
}

func TestEconomics_VerifyRewardsPerBlock_InflationZero(t *testing.T) {
	t.Parallel()

	commAddress := "protocolSustainabilityAddress"
	// based on 0.1 inflation, 14400 rounds, 3 shards, 365 epochs per year
	totalSupply := big.NewInt(100 * 4 * 14400 * 365 * 10) // 21.02B
	roundDur := 6
	args := getArguments()
	args.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func(_ uint32) float64 {
			return 0.0
		},
		ProtocolSustainabilityAddressCalled: func() string {
			return commAddress
		},
		ProtocolSustainabilityPercentageCalled: func() float64 {
			return 0.1
		},
		LeaderPercentageCalled: func() float64 {
			return 0.1
		},
	}
	args.RoundTime = &mock.RoundTimeDurationHandler{
		TimeDurationCalled: func() time.Duration {
			return time.Duration(roundDur) * time.Second
		},
	}
	newTotalSupply := big.NewInt(0).Add(totalSupply, big.NewInt(0))
	hdrPrevEpochStart := block.MetaBlock{
		Round: 0,
		Nonce: 0,
		Epoch: 0,
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:                      newTotalSupply,
				TotalToDistribute:                big.NewInt(0),
				TotalNewlyMinted:                 big.NewInt(0),
				RewardsPerBlock:                  big.NewInt(0),
				NodePrice:                        big.NewInt(10),
				RewardsForProtocolSustainability: big.NewInt(0),
			},
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, Nonce: 0},
				{ShardID: 1, Nonce: 0},
				{ShardID: 2, Nonce: 0},
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
	args.GenesisTotalSupply = totalSupply
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	numberOfShards := 4
	rewardPerBlock100 := int64(100)
	blocksPerDay := numberOfSecondsInDay / roundDur
	protocolSustainabilityRewardsForMaxBlocks := int64(numberOfSecondsInDay/roundDur) * 4 * rewardPerBlock100 / 10

	tests := []struct {
		name                                  string
		numBlocksInEpoch                      int
		accFeesInEpoch                        *big.Int
		devFeesInEpoch                        *big.Int
		expectedProtocolSustainabilityRewards *big.Int
		expectedBaseRewardPerBlock            *big.Int
		newlyMinted                           *big.Int
	}{
		{
			name:                                  "no fees",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(0),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(0),
			expectedBaseRewardPerBlock:            big.NewInt(0),
			newlyMinted:                           big.NewInt(0),
		},
		{
			name:                                  "accFees (100/block) no devFees",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(rewardPerBlock100 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(0),
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(rewardPerBlock100 - 10 - 10), // 10% protocol and 10% leaders
			newlyMinted:                           big.NewInt(0),
		},
		{
			name:                                  "accFees (100/block) and devFees (30%)",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(rewardPerBlock100 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(rewardPerBlock100 * 4 * int64(blocksPerDay) * 3 / 10), // 30% fees
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(rewardPerBlock100*7/10 - 10 - 7), // 100 - 30 (devfees) - protocol - leaders ,
			newlyMinted:                           big.NewInt(0),
		},
		{
			name:                                  "accFees (100/block) and devFees (10%)",
			numBlocksInEpoch:                      blocksPerDay,
			accFeesInEpoch:                        big.NewInt(rewardPerBlock100 * 4 * int64(blocksPerDay)),
			devFeesInEpoch:                        big.NewInt(rewardPerBlock100 * 4 * int64(blocksPerDay) * 1 / 10), // 10% fees
			expectedProtocolSustainabilityRewards: big.NewInt(protocolSustainabilityRewardsForMaxBlocks),
			expectedBaseRewardPerBlock:            big.NewInt(rewardPerBlock100*9/10 - 10 - 9),
			newlyMinted:                           big.NewInt(0),
		},
		{
			name:                                  "accFees (100/block) devFees (20%)",
			numBlocksInEpoch:                      1,
			accFeesInEpoch:                        big.NewInt(100 * 4),
			devFeesInEpoch:                        big.NewInt(100 * 4 / 5),
			expectedProtocolSustainabilityRewards: big.NewInt(40),
			expectedBaseRewardPerBlock:            big.NewInt(rewardPerBlock100*8/10 - 10 - 8),
			newlyMinted:                           big.NewInt(0),
		},
	}

	hdrPrevEpochStartHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, hdrPrevEpochStart)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalBlocksInEpoch := int64(tt.numBlocksInEpoch * numberOfShards)
			expectedTotalToDistribute := big.NewInt(tt.accFeesInEpoch.Int64())
			totalRewardsToBeDistributedPerBlock := big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(totalBlocksInEpoch))
			expectedTotalNewlyMinted := big.NewInt(0)
			expectedTotalSupply := big.NewInt(0).Add(newTotalSupply, expectedTotalNewlyMinted)
			expectedProtocolSustainabilityRewards := big.NewInt(0).Div(expectedTotalToDistribute, big.NewInt(10))
			commRewardPerBlock := big.NewInt(0).Div(expectedProtocolSustainabilityRewards, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock := big.NewInt(0).Sub(totalRewardsToBeDistributedPerBlock, commRewardPerBlock)
			feesForValidators := big.NewInt(0).Sub(tt.accFeesInEpoch, tt.devFeesInEpoch)
			feesForProposers := core.GetPercentageOfValue(feesForValidators, args.RewardsHandler.LeaderPercentage())
			accFeeRewardPerBlock := big.NewInt(0).Div(feesForProposers, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock = big.NewInt(0).Sub(adjustedRwdPerBlock, accFeeRewardPerBlock)
			devFeeRewardPerBlock := big.NewInt(0).Div(tt.devFeesInEpoch, big.NewInt(totalBlocksInEpoch))
			adjustedRwdPerBlock = big.NewInt(0).Sub(adjustedRwdPerBlock, devFeeRewardPerBlock)

			if adjustedRwdPerBlock.Cmp(big.NewInt(0)) < 0 {
				adjustedRwdPerBlock = big.NewInt(0)
			}

			mb := block.MetaBlock{
				Round: uint64(tt.numBlocksInEpoch),
				Nonce: uint64(tt.numBlocksInEpoch),
				EpochStart: block.EpochStart{
					LastFinalizedHeaders: []block.EpochStartShardData{
						{ShardID: 0, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
						{ShardID: 1, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
						{ShardID: 2, Round: uint64(tt.numBlocksInEpoch), Nonce: uint64(tt.numBlocksInEpoch)},
					},
					Economics: block.Economics{
						TotalSupply:                      expectedTotalSupply,
						TotalToDistribute:                expectedTotalToDistribute,
						TotalNewlyMinted:                 expectedTotalNewlyMinted,
						RewardsPerBlock:                  adjustedRwdPerBlock,
						NodePrice:                        big.NewInt(10),
						PrevEpochStartHash:               hdrPrevEpochStartHash,
						RewardsForProtocolSustainability: expectedProtocolSustainabilityRewards,
					},
				},
				Epoch:                  1,
				AccumulatedFeesInEpoch: tt.accFeesInEpoch,
				DevFeesInEpoch:         tt.devFeesInEpoch,
			}

			computedEconomics, err := ec.ComputeEndOfEpochEconomics(&mb)
			assert.Nil(t, err)
			assert.NotNil(t, computedEconomics)

			err = ec.VerifyRewardsPerBlock(&mb, expectedProtocolSustainabilityRewards, computedEconomics)
			assert.Nil(t, err)

			ecos, err := ec.ComputeEndOfEpochEconomics(&mb)
			assert.Nil(t, err)
			assert.True(t, expectedProtocolSustainabilityRewards.Cmp(ecos.RewardsForProtocolSustainability) == 0)

			require.Equal(t, tt.expectedProtocolSustainabilityRewards, ecos.RewardsForProtocolSustainability)
			require.Equal(t, tt.expectedBaseRewardPerBlock.Int64(), ecos.RewardsPerBlock.Int64())
			require.Equal(t, tt.newlyMinted, ecos.TotalNewlyMinted)
		})
	}
}

type testInput struct {
	blockPerEpochOneShard  uint64
	accumulatedFeesInEpoch *big.Int
	devFeesInEpoch         *big.Int
}

func TestComputeEndOfEpochEconomics(t *testing.T) {
	t.Parallel()

	totalSupply, _ := big.NewInt(0).SetString("20000000000000000000000000000", 10) // 20 Billions ERD
	nodePrice, _ := big.NewInt(0).SetString("1000000000000000000000", 10)          // 1000 ERD
	roundDuration := 4

	args := createArgsForComputeEndOfEpochEconomics(roundDuration, totalSupply, nodePrice)
	ec, _ := NewEndOfEpochEconomicsDataCreator(args)

	epochDuration := numberOfSecondsInDay
	roundsPerEpoch := uint64(epochDuration / roundDuration)

	testInputs := []testInput{
		{blockPerEpochOneShard: roundsPerEpoch, accumulatedFeesInEpoch: intToErd(1000), devFeesInEpoch: intToErd(1000)},
		{blockPerEpochOneShard: roundsPerEpoch / 2, accumulatedFeesInEpoch: intToErd(1000), devFeesInEpoch: intToErd(1000)},
		{blockPerEpochOneShard: roundsPerEpoch / 4, accumulatedFeesInEpoch: intToErd(1000), devFeesInEpoch: intToErd(1000)},
		{blockPerEpochOneShard: roundsPerEpoch / 8, accumulatedFeesInEpoch: intToErd(1000), devFeesInEpoch: intToErd(1000)},
		{blockPerEpochOneShard: roundsPerEpoch / 16, accumulatedFeesInEpoch: intToErd(1000), devFeesInEpoch: intToErd(1000)},
		{blockPerEpochOneShard: roundsPerEpoch / 32, accumulatedFeesInEpoch: intToErd(1000), devFeesInEpoch: intToErd(1000)},
		{blockPerEpochOneShard: roundsPerEpoch / 64, accumulatedFeesInEpoch: intToErd(100000000000000), devFeesInEpoch: intToErd(10000000)},
		{blockPerEpochOneShard: roundsPerEpoch, accumulatedFeesInEpoch: intToErd(100000000000000), devFeesInEpoch: intToErd(30000000000000)},
	}

	rewardsPerBlock, _ := big.NewInt(0).SetString("84559445290038908043", 10) // *based on 0.1 inflation
	for _, input := range testInputs {
		meta := &block.MetaBlock{
			AccumulatedFeesInEpoch: input.accumulatedFeesInEpoch,
			DevFeesInEpoch:         input.devFeesInEpoch,
			Epoch:                  1,
			Round:                  roundsPerEpoch,
			Nonce:                  input.blockPerEpochOneShard,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0, Round: roundsPerEpoch, Nonce: input.blockPerEpochOneShard},
					{ShardID: 1, Round: roundsPerEpoch, Nonce: input.blockPerEpochOneShard},
				},
			},
		}

		economicsBlock, err := ec.ComputeEndOfEpochEconomics(meta)
		assert.Nil(t, err)

		verifyEconomicsBlock(t, economicsBlock, input, rewardsPerBlock, nodePrice, totalSupply, roundsPerEpoch, args.RewardsHandler)
	}
}

func createArgsForComputeEndOfEpochEconomics(
	roundDuration int,
	totalSupply *big.Int,
	nodePrice *big.Int,
) ArgsNewEpochEconomics {
	commAddress := "protocolSustainabilityAddress"

	args := getArguments()
	args.RewardsHandler = &mock.RewardsHandlerStub{
		MaxInflationRateCalled: func(_ uint32) float64 {
			return 0.1
		},
		ProtocolSustainabilityAddressCalled: func() string {
			return commAddress
		},
		ProtocolSustainabilityPercentageCalled: func() float64 {
			return 0.1
		},
		LeaderPercentageCalled: func() float64 {
			return 0.1
		},
		RewardsTopUpFactorCalled: func() float64 {
			return 0.25
		},
		RewardsTopUpGradientPointCalled: func() *big.Int {
			return big.NewInt(0).Div(args.GenesisTotalSupply, big.NewInt(10))
		},
	}
	args.RoundTime = &mock.RoundTimeDurationHandler{
		TimeDurationCalled: func() time.Duration {
			return time.Duration(roundDuration) * time.Second
		},
	}
	hdrPrevEpochStart := block.MetaBlock{
		Round: 0,
		Nonce: 0,
		Epoch: 0,
		EpochStart: block.EpochStart{
			Economics: block.Economics{
				TotalSupply:                      totalSupply,
				TotalToDistribute:                big.NewInt(10),
				TotalNewlyMinted:                 big.NewInt(10),
				RewardsPerBlock:                  big.NewInt(10),
				NodePrice:                        nodePrice,
				RewardsForProtocolSustainability: big.NewInt(10),
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
	args.GenesisTotalSupply = totalSupply

	return args
}

func verifyEconomicsBlock(
	t *testing.T,
	economicsBlock *block.Economics,
	input testInput,
	rewardsPerBlock *big.Int,
	nodePrice *big.Int,
	totalSupply *big.Int,
	roundsPerEpoch uint64,
	rewardsHandler process.RewardsHandler,
) {
	totalBlocksPerEpoch := int64(input.blockPerEpochOneShard * 3)
	hitRate := float64(input.blockPerEpochOneShard) / float64(roundsPerEpoch) * 100
	printEconomicsData(economicsBlock, hitRate, totalBlocksPerEpoch)

	expectedTotalRewardsToBeDistributed := big.NewInt(0).Mul(rewardsPerBlock, big.NewInt(totalBlocksPerEpoch))
	expectedNewTokens := big.NewInt(0).Sub(expectedTotalRewardsToBeDistributed, input.accumulatedFeesInEpoch)
	if expectedNewTokens.Cmp(big.NewInt(0)) < 0 {
		expectedNewTokens = big.NewInt(0)
		expectedTotalRewardsToBeDistributed = input.accumulatedFeesInEpoch
	}

	adjustedRewardsPerBlock := big.NewInt(0).Div(expectedTotalRewardsToBeDistributed, big.NewInt(totalBlocksPerEpoch))

	// subtract developer rewards per block
	developerFeesPerBlock := big.NewInt(0).Div(input.devFeesInEpoch, big.NewInt(totalBlocksPerEpoch))
	adjustedRewardsPerBlock.Sub(adjustedRewardsPerBlock, developerFeesPerBlock)
	// subtract leader percentage per block
	rewardsForLeader := big.NewInt(0).Sub(input.accumulatedFeesInEpoch, input.devFeesInEpoch)
	rewardsForLeader = core.GetPercentageOfValue(rewardsForLeader, rewardsHandler.LeaderPercentage())
	rewardsForLeaderPerBlock := big.NewInt(0).Div(rewardsForLeader, big.NewInt(totalBlocksPerEpoch))
	adjustedRewardsPerBlock.Sub(adjustedRewardsPerBlock, rewardsForLeaderPerBlock)
	// protocolSustainabilityPercentage
	expectedProtocolSustainabilityRewards := core.GetPercentageOfValue(expectedTotalRewardsToBeDistributed, rewardsHandler.ProtocolSustainabilityPercentage())
	// subtract protocol sustainability percentage per block
	protocolSustainabilityRewardsPerBlock := big.NewInt(0).Div(expectedProtocolSustainabilityRewards, big.NewInt(totalBlocksPerEpoch))
	adjustedRewardsPerBlock.Sub(adjustedRewardsPerBlock, protocolSustainabilityRewardsPerBlock)

	assert.Equal(t, expectedNewTokens, economicsBlock.TotalNewlyMinted)
	assert.Equal(t, big.NewInt(0).Add(totalSupply, expectedNewTokens), economicsBlock.TotalSupply)
	assert.Equal(t, expectedTotalRewardsToBeDistributed, economicsBlock.TotalToDistribute)
	assert.Equal(t, expectedProtocolSustainabilityRewards, economicsBlock.RewardsForProtocolSustainability)
	assert.Equal(t, nodePrice, economicsBlock.NodePrice)
	assert.Equal(t, adjustedRewardsPerBlock, economicsBlock.RewardsPerBlock)
}

func printEconomicsData(eb *block.Economics, hitRate float64, numBlocksTotal int64) {
	fmt.Printf("Hit rate per shard %.4f%%, Total block produced: %d \n", hitRate, numBlocksTotal)
	fmt.Printf("Total supply: %vERD, TotalToDistribute %vERD, "+
		"TotalNewlyMinted %vERD, RewardsPerBlock %vERD, RewardsForProtocolSustainability %vERD, NodePrice: %vERD\n",
		denomination(eb.TotalSupply), denomination(eb.TotalToDistribute), denomination(eb.TotalNewlyMinted),
		denomination(eb.RewardsPerBlock), denomination(eb.RewardsForProtocolSustainability), denomination(eb.NodePrice))
}

func intToErd(value int) *big.Int {
	denom, _ := big.NewInt(0).SetString("1000000000000000000", 10)

	return big.NewInt(0).Mul(denom, big.NewInt(int64(value)))
}

func denomination(value *big.Int) string {
	denom, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	cpValue := big.NewInt(0).Set(value)

	return cpValue.Div(cpValue, denom).String()
}

func getArguments() ArgsNewEpochEconomics {
	return ArgsNewEpochEconomics{
		Marshalizer:           &mock.MarshalizerMock{},
		Hasher:                mock.HasherMock{},
		Store:                 &mock.ChainStorerStub{},
		ShardCoordinator:      mock.NewMultipleShardsCoordinatorMock(),
		RewardsHandler:        &mock.RewardsHandlerStub{},
		RoundTime:             &mock.RoundTimeDurationHandler{},
		GenesisTotalSupply:    big.NewInt(200000000),
		EconomicsDataNotified: NewEpochEconomicsStatistics(),
	}
}
