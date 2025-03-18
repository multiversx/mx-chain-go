package rating

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	shardValidatorIncreaseRatingStep = int32(2)
	shardValidatorDecreaseRatingStep = int32(-8)
	shardProposerIncreaseRatingStep  = int32(6)
	shardProposerDecreaseRatingStep  = int32(-24)

	metaValidatorIncreaseRatingStep = int32(1)
	metaValidatorDecreaseRatingStep = int32(-4)
	metaProposerIncreaseRatingStep  = int32(6)
	metaProposerDecreaseRatingStep  = int32(-24)

	signedBlocksThreshold          = 0.025
	consecutiveMissedBlocksPenalty = 1.1

	shardMinNodes             = 6
	shardConsensusSize        = 3
	metaMinNodes              = 6
	metaConsensusSize         = 6
	roundDurationMilliseconds = 6000
)

func createDummyRatingsData() RatingsDataArg {
	return RatingsDataArg{
		Config: config.RatingsConfig{},
		ChainParametersHolder: &chainParameters.ChainParametersHandlerStub{
			CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
				return config.ChainParametersByEpochConfig{
					RoundDuration:               4000,
					Hysteresis:                  0.2,
					EnableEpoch:                 0,
					ShardConsensusGroupSize:     shardConsensusSize,
					ShardMinNumNodes:            shardMinNodes,
					MetachainConsensusGroupSize: metaConsensusSize,
					MetachainMinNumNodes:        metaMinNodes,
					Adaptivity:                  false,
				}
			},
			AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
				return []config.ChainParametersByEpochConfig{
					{
						RoundDuration:               4000,
						Hysteresis:                  0.2,
						EnableEpoch:                 0,
						ShardConsensusGroupSize:     shardConsensusSize,
						ShardMinNumNodes:            shardMinNodes,
						MetachainConsensusGroupSize: metaConsensusSize,
						MetachainMinNumNodes:        metaMinNodes,
						Adaptivity:                  false,
					},
				}
			},
		},
		RoundDurationMilliseconds: roundDurationMilliseconds,
		EpochNotifier:             &epochNotifier.EpochNotifierStub{},
	}
}

func createDummyRatingsConfig() config.RatingsConfig {
	return config.RatingsConfig{
		General: config.General{
			StartRating:           4000,
			MaxRating:             10000,
			MinRating:             1,
			SignedBlocksThreshold: signedBlocksThreshold,
			SelectionChances: []*config.SelectionChance{
				{MaxThreshold: 0, ChancePercent: 5},
				{MaxThreshold: 25, ChancePercent: 19},
				{MaxThreshold: 75, ChancePercent: 20},
				{MaxThreshold: 100, ChancePercent: 21},
			},
		},
		ShardChain: config.ShardChain{
			RatingStepsByEpoch: []config.RatingSteps{
				{
					HoursToMaxRatingFromStartRating: 2,
					ProposerValidatorImportance:     1,
					ProposerDecreaseFactor:          -4,
					ValidatorDecreaseFactor:         -4,
					ConsecutiveMissedBlocksPenalty:  consecutiveMissedBlocksPenalty,
					EnableEpoch:                     0,
				},
			},
		},
		MetaChain: config.MetaChain{
			RatingStepsByEpoch: []config.RatingSteps{
				{
					HoursToMaxRatingFromStartRating: 2,
					ProposerValidatorImportance:     1,
					ProposerDecreaseFactor:          -4,
					ValidatorDecreaseFactor:         -4,
					ConsecutiveMissedBlocksPenalty:  consecutiveMissedBlocksPenalty,
					EnableEpoch:                     0,
				},
			},
		},
	}
}

func TestNewRatingsData_NilEpochNotifier(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsDataArg.EpochNotifier = nil

	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrNilEpochNotifier))
}

func TestNewRatingsData_NilChainParametersHolder(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsDataArg.ChainParametersHolder = nil

	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrNilChainParametersHandler))
}

func TestNewRatingsData_MissingConfigurationForEpoch0(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsDataArg.Config = createDummyRatingsConfig()
	ratingsDataArg.Config.ShardChain.RatingStepsByEpoch[0].EnableEpoch = 37
	ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].EnableEpoch = 37
	ratingsDataArg.ChainParametersHolder = &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{
				RoundDuration:               4000,
				Hysteresis:                  0.2,
				EnableEpoch:                 37,
				ShardConsensusGroupSize:     shardConsensusSize,
				ShardMinNumNodes:            shardMinNodes,
				MetachainConsensusGroupSize: metaConsensusSize,
				MetachainMinNumNodes:        metaMinNodes,
				Adaptivity:                  false,
			}
		},
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return []config.ChainParametersByEpochConfig{
				{
					RoundDuration:               4000,
					Hysteresis:                  0.2,
					EnableEpoch:                 37,
					ShardConsensusGroupSize:     shardConsensusSize,
					ShardMinNumNodes:            shardMinNodes,
					MetachainConsensusGroupSize: metaConsensusSize,
					MetachainMinNumNodes:        metaMinNodes,
					Adaptivity:                  false,
				},
			}
		},
	}

	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrMissingConfigurationForEpochZero))
}

func TestRatingsData_RatingsDataMinGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 10
	ratingsConfig.General.MaxRating = 8

	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrMaxRatingIsSmallerThanMinRating))
}

func TestRatingsData_RatingsDataMinSmallerThanOne(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 0
	ratingsConfig.General.MaxRating = 8
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.Equal(t, process.ErrMinRatingSmallerThanOne, err)
}

func TestRatingsData_RatingsStartGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 10
	ratingsConfig.General.MaxRating = 100
	ratingsConfig.General.StartRating = 110
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrStartRatingNotBetweenMinAndMax))
}

func TestRatingsData_RatingsStartLowerMinShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 10
	ratingsConfig.General.MaxRating = 100
	ratingsConfig.General.StartRating = 5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrStartRatingNotBetweenMinAndMax))
}

func TestRatingsData_RatingsSignedBlocksThresholdNotBetweenZeroAndOneShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.SignedBlocksThreshold = -0.1
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne))

	ratingsConfig.General.SignedBlocksThreshold = 1.01
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne))
}

func TestRatingsData_RatingsConsecutiveMissedBlocksPenaltyLowerThanOneShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty = 0.9
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty = 1.99
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty = 0.99
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "shard"))
}

func TestRatingsData_EmptyRatingsConfig(t *testing.T) {
	t.Parallel()

	t.Run("shard should error", func(t *testing.T) {
		t.Parallel()

		ratingsDataArg := createDummyRatingsData()
		ratingsConfig := createDummyRatingsConfig()
		ratingsConfig.ShardChain = config.ShardChain{}
		ratingsDataArg.Config = ratingsConfig
		ratingsData, err := NewRatingsData(ratingsDataArg)

		require.Nil(t, ratingsData)
		require.True(t, errors.Is(err, process.ErrInvalidRatingsConfig))
		require.True(t, strings.Contains(err.Error(), "shardChain"))
	})
	t.Run("meta should error", func(t *testing.T) {
		t.Parallel()

		ratingsDataArg := createDummyRatingsData()
		ratingsConfig := createDummyRatingsConfig()
		ratingsConfig.MetaChain = config.MetaChain{}
		ratingsDataArg.Config = ratingsConfig
		ratingsData, err := NewRatingsData(ratingsDataArg)

		require.Nil(t, ratingsData)
		require.True(t, errors.Is(err, process.ErrInvalidRatingsConfig))
		require.True(t, strings.Contains(err.Error(), "metaChain"))
	})
}

func TestRatingsData_HoursToMaxRatingFromStartRatingZeroErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].HoursToMaxRatingFromStartRating = 0
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrHoursToMaxRatingFromStartRatingZero))
}

func TestRatingsData_PositiveDecreaseRatingsStepsShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ProposerDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ValidatorDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ProposerDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "shard"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ValidatorDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "shard"))
}

func TestRatingsData_UnderflowErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ProposerDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerDecrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ValidatorDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorDecrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ProposerDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerDecrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ValidatorDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorDecrease"))
}

func TestRatingsData_EpochConfirmed(t *testing.T) {
	t.Parallel()

	// Activation epochs for this test:
	// 		0 -> new chain params but same values as epoch 0, new ratingSteps for shard, new ratingSteps for meta
	// 		4 -> same chain params as epoch 0, new ratingSteps for shard, same meta ratingSteps as epoch 0
	// 		5 -> new chain params, same shard ratingSteps as epoch 4, same meta ratingSteps as epoch 0
	// 		7 -> same chain params as epoch 5, new shard ratingSteps, new meta ratingSteps
	// 	   10 -> new chain params but same values as epoch 5, same shard ratingSteps as epoch 7, same meta ratingSteps as epoch 7
	// 	   15 -> new chain params, new shard ratingSteps, new meta ratingSteps
	chainParams := make([]config.ChainParametersByEpochConfig, 0)
	for i := uint32(0); i <= 15; i += 5 {
		newChainParams := config.ChainParametersByEpochConfig{
			RoundDuration:               4000,
			Hysteresis:                  0.2,
			EnableEpoch:                 i,
			ShardConsensusGroupSize:     shardConsensusSize,
			ShardMinNumNodes:            shardMinNodes,
			MetachainConsensusGroupSize: metaConsensusSize,
			MetachainMinNumNodes:        metaMinNodes,
			Adaptivity:                  false,
		}
		// change consensus size for shard after epoch 5
		if i >= 5 {
			newChainParams.ShardConsensusGroupSize = shardConsensusSize + i
		}

		chainParams = append(chainParams, newChainParams)
	}
	expectedChainParamsIdx := 0
	chainParamsHandler := &chainParameters.ChainParametersHandlerStub{
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return chainParams
		},
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return chainParams[expectedChainParamsIdx]
		},
	}
	ratingsDataArg := createDummyRatingsData()
	ratingsDataArg.Config = createDummyRatingsConfig()
	ratingsDataArg.Config.ShardChain.RatingStepsByEpoch = []config.RatingSteps{
		{
			HoursToMaxRatingFromStartRating: 1,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.1,
			EnableEpoch:                     0,
		},
		{
			HoursToMaxRatingFromStartRating: 2,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.5,
			EnableEpoch:                     4,
		},
		{
			HoursToMaxRatingFromStartRating: 2,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.7,
			EnableEpoch:                     7,
		},
		{
			HoursToMaxRatingFromStartRating: 1,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.8,
			EnableEpoch:                     15,
		},
	}
	ratingsDataArg.Config.MetaChain.RatingStepsByEpoch = []config.RatingSteps{
		{
			HoursToMaxRatingFromStartRating: 2,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.5,
			EnableEpoch:                     0,
		},
		{
			HoursToMaxRatingFromStartRating: 2,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.7,
			EnableEpoch:                     7,
		},
		{
			HoursToMaxRatingFromStartRating: 1,
			ProposerValidatorImportance:     1,
			ProposerDecreaseFactor:          -4,
			ValidatorDecreaseFactor:         -4,
			ConsecutiveMissedBlocksPenalty:  1.9,
			EnableEpoch:                     15,
		},
	}
	ratingsDataArg.ChainParametersHolder = chainParamsHandler
	rd, err := NewRatingsData(ratingsDataArg)
	require.NoError(t, err)
	require.NotNil(t, rd)

	cntSetInt64ValueHandler := 0
	cntSetStringValueHandler := 0
	handler := &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			cntSetInt64ValueHandler++
		},
		SetStringValueHandler: func(key string, value string) {
			cntSetStringValueHandler++
		},
	}
	_ = rd.SetStatusHandler(handler)

	// ensure that the configs are stored in descending order
	currentConfig := rd.ratingsStepsConfig[0]
	for i := 1; i < len(rd.ratingsStepsConfig); i++ {
		require.Less(t, rd.ratingsStepsConfig[i].enableEpoch, currentConfig.enableEpoch)
		currentConfig = rd.ratingsStepsConfig[i]
	}

	// check epoch 0
	require.Equal(t, uint32(0), rd.currentRatingsStepData.enableEpoch)
	expectedConsecutiveMissedBlocksPenaltyShardEpoch0 := ratingsDataArg.Config.ShardChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch0, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	expectedConsecutiveMissedBlocksPenaltyMetaEpoch0 := ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch0, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 1, nothing changed, same as before
	rd.EpochConfirmed(1, 0)
	require.Equal(t, uint32(0), rd.currentRatingsStepData.enableEpoch)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch0, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch0, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 4, shard changed, chain params changed
	rd.EpochConfirmed(4, 0)
	require.Equal(t, uint32(4), rd.currentRatingsStepData.enableEpoch)
	expectedConsecutiveMissedBlocksPenaltyShardEpoch4 := ratingsDataArg.Config.ShardChain.RatingStepsByEpoch[1].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch4, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	expectedConsecutiveMissedBlocksPenaltyMetaEpoch4 := ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch4, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 5, nothing changed, same as before, but we have new chain params defined for this epoch
	rd.EpochConfirmed(5, 0)
	require.Equal(t, uint32(5), rd.currentRatingsStepData.enableEpoch)
	expectedChainParamsIdx = 1 // epoch 5
	require.Equal(t, uint32(shardConsensusSize+5), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch4, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch4, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 6, nothing changed, same as before
	rd.EpochConfirmed(6, 0)
	require.Equal(t, uint32(5), rd.currentRatingsStepData.enableEpoch)
	require.Equal(t, uint32(shardConsensusSize+5), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch4, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch4, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 7, shard changed, meta changed, same chain params
	rd.EpochConfirmed(7, 0)
	require.Equal(t, uint32(7), rd.currentRatingsStepData.enableEpoch)
	require.Equal(t, uint32(shardConsensusSize+5), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	expectedConsecutiveMissedBlocksPenaltyShardEpoch7 := ratingsDataArg.Config.ShardChain.RatingStepsByEpoch[2].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch7, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	expectedConsecutiveMissedBlocksPenaltyMetaEpoch7 := ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[1].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch7, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 9, nothing changed, same as before
	rd.EpochConfirmed(9, 0)
	require.Equal(t, uint32(7), rd.currentRatingsStepData.enableEpoch)
	require.Equal(t, uint32(shardConsensusSize+5), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch7, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch7, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 10, nothing changed, same as before, but we have new chain params defined for this epoch
	rd.EpochConfirmed(10, 0)
	require.Equal(t, uint32(10), rd.currentRatingsStepData.enableEpoch)
	expectedChainParamsIdx = 2 // epoch 10
	require.Equal(t, uint32(shardConsensusSize+10), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch7, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch7, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 11, nothing changed, same as before
	rd.EpochConfirmed(11, 0)
	require.Equal(t, uint32(10), rd.currentRatingsStepData.enableEpoch)
	require.Equal(t, uint32(shardConsensusSize+10), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch7, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch7, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 15, shard changed, meta changed, chain params changed
	rd.EpochConfirmed(15, 0)
	require.Equal(t, uint32(15), rd.currentRatingsStepData.enableEpoch)
	expectedChainParamsIdx = 3 // epoch 15
	require.Equal(t, uint32(shardConsensusSize+15), rd.chainParametersHandler.CurrentChainParameters().ShardConsensusGroupSize)
	expectedConsecutiveMissedBlocksPenaltyShardEpoch15 := ratingsDataArg.Config.ShardChain.RatingStepsByEpoch[3].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch15, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	expectedConsecutiveMissedBlocksPenaltyMetaEpoch15 := ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[2].ConsecutiveMissedBlocksPenalty
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch15, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	// check epoch 429, nothing changed, same as before
	rd.EpochConfirmed(429, 0)
	require.Equal(t, uint32(15), rd.currentRatingsStepData.enableEpoch)
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyShardEpoch15, rd.currentRatingsStepData.shardRatingsStepData.ConsecutiveMissedBlocksPenalty())
	require.Equal(t, expectedConsecutiveMissedBlocksPenaltyMetaEpoch15, rd.currentRatingsStepData.metaRatingsStepData.ConsecutiveMissedBlocksPenalty())

	expectedNumberOfConfigChanges := 5
	require.Equal(t, expectedNumberOfConfigChanges*2, cntSetInt64ValueHandler)  // for each epoch confirmed should be called twice, shard + meta
	require.Equal(t, expectedNumberOfConfigChanges*8, cntSetStringValueHandler) // for each epoch confirmed should be called 8 times, 4 for shard, 4 for meta
}

func TestRatingsData_OverflowErr(t *testing.T) {
	t.Parallel()

	getBaseChainParams := func() config.ChainParametersByEpochConfig {
		return config.ChainParametersByEpochConfig{
			RoundDuration:               4000,
			Hysteresis:                  0.2,
			EnableEpoch:                 0,
			ShardConsensusGroupSize:     5,
			ShardMinNumNodes:            7,
			MetachainConsensusGroupSize: 7,
			MetachainMinNumNodes:        7,
			Adaptivity:                  false,
		}
	}
	getChainParametersHandler := func(cfg config.ChainParametersByEpochConfig) *chainParameters.ChainParametersHandlerStub {
		return &chainParameters.ChainParametersHandlerStub{
			CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
				return cfg
			},
			AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
				return []config.ChainParametersByEpochConfig{cfg}
			},
		}
	}

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	chainParams := getBaseChainParams()
	chainParams.RoundDuration = 3600 * 1000
	chainParams.MetachainMinNumNodes = math.MaxUint32
	ratingsDataArg.ChainParametersHolder = getChainParametersHandler(chainParams)
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	chainParams = getBaseChainParams()
	chainParams.RoundDuration = 3600 * 1000
	chainParams.MetachainMinNumNodes = math.MaxUint32
	chainParams.MetachainConsensusGroupSize = 1
	ratingsDataArg.ChainParametersHolder = getChainParametersHandler(chainParams)
	ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].ProposerValidatorImportance = float32(1) / math.MaxUint32
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorIncrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	chainParams = getBaseChainParams()
	chainParams.RoundDuration = 3600 * 1000
	chainParams.ShardMinNumNodes = math.MaxUint32
	ratingsDataArg.ChainParametersHolder = getChainParametersHandler(chainParams)
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	chainParams = getBaseChainParams()
	chainParams.RoundDuration = 3600 * 1000
	chainParams.ShardMinNumNodes = math.MaxUint32
	chainParams.ShardConsensusGroupSize = 1
	ratingsDataArg.ChainParametersHolder = getChainParametersHandler(chainParams)
	ratingsDataArg.Config.ShardChain.RatingStepsByEpoch[0].ProposerValidatorImportance = float32(1) / math.MaxUint32
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorIncrease"))
}

func TestRatingsData_IncreaseLowerThanZeroErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].HoursToMaxRatingFromStartRating = math.MaxUint32
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrIncreaseStepLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].HoursToMaxRatingFromStartRating = 2
	ratingsDataArg.Config.MetaChain.RatingStepsByEpoch[0].ProposerValidatorImportance = math.MaxUint32
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrIncreaseStepLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "validatorIncrease"))
}

func TestRatingsData_RatingsCorrectValues(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	minRating := uint32(1)
	maxRating := uint32(10000)
	startRating := uint32(4000)
	signedBlocksThreshold := float32(0.025)
	shardConsecutivePenalty := float32(1.2)
	metaConsecutivePenalty := float32(1.3)
	hoursToMaxRatingFromStartRating := uint32(5)
	decreaseFactor := float32(-4)

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = minRating
	ratingsConfig.General.MaxRating = maxRating
	ratingsConfig.General.StartRating = startRating
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].HoursToMaxRatingFromStartRating = hoursToMaxRatingFromStartRating
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].HoursToMaxRatingFromStartRating = hoursToMaxRatingFromStartRating
	ratingsConfig.General.SignedBlocksThreshold = signedBlocksThreshold
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty = shardConsecutivePenalty
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ProposerDecreaseFactor = decreaseFactor
	ratingsConfig.ShardChain.RatingStepsByEpoch[0].ValidatorDecreaseFactor = decreaseFactor
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ConsecutiveMissedBlocksPenalty = metaConsecutivePenalty
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ProposerDecreaseFactor = decreaseFactor
	ratingsConfig.MetaChain.RatingStepsByEpoch[0].ValidatorDecreaseFactor = decreaseFactor

	selectionChances := []*config.SelectionChance{
		{MaxThreshold: 0, ChancePercent: 1},
		{MaxThreshold: minRating, ChancePercent: 2},
		{MaxThreshold: maxRating, ChancePercent: 4},
	}

	ratingsConfig.General.SelectionChances = selectionChances

	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	assert.Nil(t, err)
	assert.NotNil(t, ratingsData)
	assert.Equal(t, startRating, ratingsData.StartRating())
	assert.Equal(t, minRating, ratingsData.MinRating())
	assert.Equal(t, maxRating, ratingsData.MaxRating())
	assert.Equal(t, signedBlocksThreshold, ratingsData.SignedBlocksThreshold())
	assert.Equal(t, shardValidatorIncreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ValidatorIncreaseRatingStep())
	assert.Equal(t, shardValidatorDecreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	assert.Equal(t, shardProposerIncreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ProposerIncreaseRatingStep())
	assert.Equal(t, shardProposerDecreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ProposerDecreaseRatingStep())
	assert.Equal(t, metaValidatorIncreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ValidatorIncreaseRatingStep())
	assert.Equal(t, metaValidatorDecreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	assert.Equal(t, metaProposerIncreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ProposerIncreaseRatingStep())
	assert.Equal(t, metaProposerDecreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ProposerDecreaseRatingStep())
	assert.Equal(t, shardConsecutivePenalty, ratingsData.ShardChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty())
	assert.Equal(t, metaConsecutivePenalty, ratingsData.MetaChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty())

	for i := range selectionChances {
		assert.Equal(t, selectionChances[i].MaxThreshold, ratingsData.SelectionChances()[i].GetMaxThreshold())
		assert.Equal(t, selectionChances[i].ChancePercent, ratingsData.SelectionChances()[i].GetChancePercent())
	}
}

func TestRatingsData_SetStatusHandler(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsDataArg.Config = createDummyRatingsConfig()
	ratingsData, _ := NewRatingsData(ratingsDataArg)
	require.NotNil(t, ratingsData)

	err := ratingsData.SetStatusHandler(nil)
	require.Equal(t, process.ErrNilAppStatusHandler, err)

	err = ratingsData.SetStatusHandler(&statusHandler.AppStatusHandlerStub{})
	require.NoError(t, err)
}
