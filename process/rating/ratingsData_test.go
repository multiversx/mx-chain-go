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
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 2,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  consecutiveMissedBlocksPenalty,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				HoursToMaxRatingFromStartRating: 2,
				ProposerValidatorImportance:     1,
				ProposerDecreaseFactor:          -4,
				ValidatorDecreaseFactor:         -4,
				ConsecutiveMissedBlocksPenalty:  consecutiveMissedBlocksPenalty,
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
	ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty = 0.9
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty = 1.99
	ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty = 0.99
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "shard"))
}

func TestRatingsData_HoursToMaxRatingFromStartRatingZeroErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.HoursToMaxRatingFromStartRating = 0
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrHoursToMaxRatingFromStartRatingZero))
}

func TestRatingsData_PositiveDecreaseRatingsStepsShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDummyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.ProposerDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.MetaChain.ValidatorDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ProposerDecreaseFactor = -0.5
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepMoreThanMinusOne))
	require.True(t, strings.Contains(err.Error(), "shard"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ValidatorDecreaseFactor = -0.5
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
	ratingsConfig.MetaChain.ProposerDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerDecrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.MetaChain.ValidatorDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorDecrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ProposerDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerDecrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ValidatorDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorDecrease"))
}

func TestRatingsData_EpochConfirmed(t *testing.T) {
	t.Parallel()

	chainParams := make([]config.ChainParametersByEpochConfig, 0)
	for i := uint32(0); i <= 10; i += 5 {
		chainParams = append(chainParams, config.ChainParametersByEpochConfig{
			RoundDuration:               4000,
			Hysteresis:                  0.2,
			EnableEpoch:                 i,
			ShardConsensusGroupSize:     shardConsensusSize,
			ShardMinNumNodes:            shardMinNodes,
			MetachainConsensusGroupSize: metaConsensusSize,
			MetachainMinNumNodes:        metaMinNodes,
			Adaptivity:                  false,
		})
	}
	chainParamsHandler := &chainParameters.ChainParametersHandlerStub{
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return chainParams
		},
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return chainParams[0]
		},
	}
	ratingsDataArg := createDummyRatingsData()
	ratingsDataArg.Config = createDummyRatingsConfig()
	ratingsDataArg.ChainParametersHolder = chainParamsHandler
	rd, err := NewRatingsData(ratingsDataArg)
	require.NoError(t, err)
	require.NotNil(t, rd)

	// ensure that the configs are stored in descending order
	currentConfig := rd.ratingsStepsConfig[0]
	for i := 1; i < len(rd.ratingsStepsConfig); i++ {
		require.Less(t, rd.ratingsStepsConfig[i].enableEpoch, currentConfig.enableEpoch)
		currentConfig = rd.ratingsStepsConfig[i]
	}

	require.Equal(t, uint32(0), rd.currentRatingsStepData.enableEpoch)

	rd.EpochConfirmed(4, 0)
	require.Equal(t, uint32(0), rd.currentRatingsStepData.enableEpoch)

	rd.EpochConfirmed(5, 0)
	require.Equal(t, uint32(5), rd.currentRatingsStepData.enableEpoch)

	rd.EpochConfirmed(9, 0)
	require.Equal(t, uint32(5), rd.currentRatingsStepData.enableEpoch)

	rd.EpochConfirmed(10, 0)
	require.Equal(t, uint32(10), rd.currentRatingsStepData.enableEpoch)

	rd.EpochConfirmed(11, 0)
	require.Equal(t, uint32(10), rd.currentRatingsStepData.enableEpoch)

	rd.EpochConfirmed(429, 0)
	require.Equal(t, uint32(10), rd.currentRatingsStepData.enableEpoch)
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
	ratingsDataArg.Config.MetaChain.ProposerValidatorImportance = float32(1) / math.MaxUint32
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
	ratingsDataArg.Config.ShardChain.ProposerValidatorImportance = float32(1) / math.MaxUint32
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
	ratingsDataArg.Config.MetaChain.HoursToMaxRatingFromStartRating = math.MaxUint32
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrIncreaseStepLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDummyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.Config.MetaChain.HoursToMaxRatingFromStartRating = 2
	ratingsDataArg.Config.MetaChain.ProposerValidatorImportance = math.MaxUint32
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
	ratingsConfig.MetaChain.HoursToMaxRatingFromStartRating = hoursToMaxRatingFromStartRating
	ratingsConfig.ShardChain.HoursToMaxRatingFromStartRating = hoursToMaxRatingFromStartRating
	ratingsConfig.General.SignedBlocksThreshold = signedBlocksThreshold
	ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty = shardConsecutivePenalty
	ratingsConfig.ShardChain.ProposerDecreaseFactor = decreaseFactor
	ratingsConfig.ShardChain.ValidatorDecreaseFactor = decreaseFactor
	ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty = metaConsecutivePenalty
	ratingsConfig.MetaChain.ProposerDecreaseFactor = decreaseFactor
	ratingsConfig.MetaChain.ValidatorDecreaseFactor = decreaseFactor

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
