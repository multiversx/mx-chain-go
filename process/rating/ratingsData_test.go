package rating

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
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

	shardMinNodes            = 6
	shardConsensusSize       = 3
	metaMinNodes             = 6
	metaConsensusSize        = 6
	roundDurationMiliseconds = 6000
)

func createDymmyRatingsData() RatingsDataArg {
	return RatingsDataArg{
		Config:                   config.RatingsConfig{},
		ShardConsensusSize:       shardConsensusSize,
		MetaConsensusSize:        metaConsensusSize,
		ShardMinNodes:            shardMinNodes,
		MetaMinNodes:             metaMinNodes,
		RoundDurationMiliseconds: roundDurationMiliseconds,
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

func TestRatingsData_RatingsDataMinGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.HoursToMaxRatingFromStartRating = 0
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrHoursToMaxRatingFromStartRatingZero))
}

func TestRatingsData_PositiveDecreaseRatingsStepsShouldErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.ProposerDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerDecrease"))

	ratingsDataArg = createDymmyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.MetaChain.ValidatorDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorDecrease"))

	ratingsDataArg = createDymmyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ProposerDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerDecrease"))

	ratingsDataArg = createDymmyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ValidatorDecreaseFactor = math.MinInt32
	ratingsDataArg.Config = ratingsConfig
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorDecrease"))
}

func TestRatingsData_OverflowErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDymmyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.RoundDurationMiliseconds = 3600 * 1000
	ratingsDataArg.MetaMinNodes = math.MaxUint32
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDymmyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.RoundDurationMiliseconds = 3600 * 1000
	ratingsDataArg.MetaMinNodes = math.MaxUint32
	ratingsDataArg.MetaConsensusSize = 1
	ratingsDataArg.Config.MetaChain.ProposerValidatorImportance = float32(1) / math.MaxUint32
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorIncrease"))

	ratingsDataArg = createDymmyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.RoundDurationMiliseconds = 3600 * 1000
	ratingsDataArg.ShardMinNodes = math.MaxUint32
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDymmyRatingsData()
	ratingsConfig = createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.RoundDurationMiliseconds = 3600 * 1000
	ratingsDataArg.ShardMinNodes = math.MaxUint32
	ratingsDataArg.ShardConsensusSize = 1
	ratingsDataArg.Config.ShardChain.ProposerValidatorImportance = float32(1) / math.MaxUint32
	ratingsData, err = NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrOverflow))
	require.True(t, strings.Contains(err.Error(), "validatorIncrease"))
}

func TestRatingsData_IncreaseLowerThanZeroErr(t *testing.T) {
	t.Parallel()

	ratingsDataArg := createDymmyRatingsData()
	ratingsConfig := createDummyRatingsConfig()
	ratingsDataArg.Config = ratingsConfig
	ratingsDataArg.Config.MetaChain.HoursToMaxRatingFromStartRating = math.MaxUint32
	ratingsData, err := NewRatingsData(ratingsDataArg)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrIncreaseStepLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "proposerIncrease"))

	ratingsDataArg = createDymmyRatingsData()
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

	ratingsDataArg := createDymmyRatingsData()
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
