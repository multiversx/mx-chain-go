package rating

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	shardValidatorIncreaseRatingStep = int32(2)
	shardValidatorDecreaseRatingStep = int32(-4)
	shardProposerIncreaseRatingStep  = int32(1)
	shardProposerDecreaseRatingStep  = int32(-2)

	metaValidatorIncreaseRatingStep = int32(2)
	metaValidatorDecreaseRatingStep = int32(-4)
	metaProposerIncreaseRatingStep  = int32(1)
	metaProposerDecreaseRatingStep  = int32(-2)

	signedBlocksThreshold          = 0.025
	consecutiveMissedBlocksPenalty = 1.1
)

func createDummyRatingsConfig() config.RatingsConfig {
	return config.RatingsConfig{
		General: config.General{
			StartRating:           50,
			MaxRating:             100,
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
				ProposerIncreaseRatingStep:     shardProposerIncreaseRatingStep,
				ProposerDecreaseRatingStep:     shardProposerDecreaseRatingStep,
				ValidatorIncreaseRatingStep:    shardValidatorIncreaseRatingStep,
				ValidatorDecreaseRatingStep:    shardValidatorDecreaseRatingStep,
				ConsecutiveMissedBlocksPenalty: consecutiveMissedBlocksPenalty,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				ProposerIncreaseRatingStep:     metaProposerIncreaseRatingStep,
				ProposerDecreaseRatingStep:     metaProposerDecreaseRatingStep,
				ValidatorIncreaseRatingStep:    metaValidatorIncreaseRatingStep,
				ValidatorDecreaseRatingStep:    metaValidatorDecreaseRatingStep,
				ConsecutiveMissedBlocksPenalty: consecutiveMissedBlocksPenalty,
			},
		},
	}
}

func TestRatingsData_RatingsDataMinGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 10
	ratingsConfig.General.MaxRating = 8
	ratingsData, err := NewRatingsData(ratingsConfig)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrMaxRatingIsSmallerThanMinRating))
}

func TestRatingsData_RatingsDataMinSmallerThanOne(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 0
	ratingsConfig.General.MaxRating = 8
	ratingsData, err := NewRatingsData(ratingsConfig)

	assert.Nil(t, ratingsData)
	assert.Equal(t, process.ErrMinRatingSmallerThanOne, err)
}

func TestRatingsData_RatingsStartGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 10
	ratingsConfig.General.MaxRating = 100
	ratingsConfig.General.StartRating = 110
	ratingsData, err := NewRatingsData(ratingsConfig)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrStartRatingNotBetweenMinAndMax))
}

func TestRatingsData_RatingsStartLowerMinShouldErr(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = 10
	ratingsConfig.General.MaxRating = 100
	ratingsConfig.General.StartRating = 5
	ratingsData, err := NewRatingsData(ratingsConfig)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrStartRatingNotBetweenMinAndMax))
}

func TestRatingsData_RatingsSignedBlocksThresholdNotBetweenZeroAndOneShouldErr(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.SignedBlocksThreshold = -0.1
	ratingsData, err := NewRatingsData(ratingsConfig)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne))

	ratingsConfig.General.SignedBlocksThreshold = 1.01
	ratingsData, err = NewRatingsData(ratingsConfig)

	assert.Nil(t, ratingsData)
	assert.True(t, errors.Is(err, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne))
}

func TestRatingsData_RatingsConsecutiveMissedBlocksPenaltyLowerThanOneShouldErr(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty = 0.9
	ratingsData, err := NewRatingsData(ratingsConfig)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty = 1.99
	ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty = 0.99
	ratingsData, err = NewRatingsData(ratingsConfig)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne))
	require.True(t, strings.Contains(err.Error(), "shard"))
}

func TestRatingsData_PositiveDecreaseRatingsStepsShouldErr(t *testing.T) {
	t.Parallel()

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.MetaChain.ProposerDecreaseRatingStep = 7
	ratingsData, err := NewRatingsData(ratingsConfig)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepPositive))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.MetaChain.ValidatorDecreaseRatingStep = 7
	ratingsData, err = NewRatingsData(ratingsConfig)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepPositive))
	require.True(t, strings.Contains(err.Error(), "meta"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ProposerDecreaseRatingStep = 7
	ratingsData, err = NewRatingsData(ratingsConfig)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepPositive))
	require.True(t, strings.Contains(err.Error(), "shard"))

	ratingsConfig = createDummyRatingsConfig()
	ratingsConfig.ShardChain.ValidatorDecreaseRatingStep = 7
	ratingsData, err = NewRatingsData(ratingsConfig)

	require.Nil(t, ratingsData)
	require.True(t, errors.Is(err, process.ErrDecreaseRatingsStepPositive))
	require.True(t, strings.Contains(err.Error(), "shard"))
}

func TestRatingsData_RatingsCorrectValues(t *testing.T) {
	t.Parallel()

	minRating := uint32(10)
	maxRating := uint32(100)
	startRating := uint32(50)
	signedBlocksThreshold := float32(0.025)
	shardConsecutivePenalty := float32(1.2)
	metaConsecutivePenalty := float32(1.3)
	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = minRating
	ratingsConfig.General.MaxRating = maxRating
	ratingsConfig.General.StartRating = startRating
	ratingsConfig.General.SignedBlocksThreshold = signedBlocksThreshold
	ratingsConfig.ShardChain.ProposerDecreaseRatingStep = shardProposerDecreaseRatingStep
	ratingsConfig.ShardChain.ProposerIncreaseRatingStep = shardProposerIncreaseRatingStep
	ratingsConfig.ShardChain.ValidatorIncreaseRatingStep = shardValidatorIncreaseRatingStep
	ratingsConfig.ShardChain.ValidatorDecreaseRatingStep = shardValidatorDecreaseRatingStep
	ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty = shardConsecutivePenalty
	ratingsConfig.MetaChain.ProposerDecreaseRatingStep = metaProposerDecreaseRatingStep
	ratingsConfig.MetaChain.ProposerIncreaseRatingStep = metaProposerIncreaseRatingStep
	ratingsConfig.MetaChain.ValidatorIncreaseRatingStep = metaValidatorIncreaseRatingStep
	ratingsConfig.MetaChain.ValidatorDecreaseRatingStep = metaValidatorDecreaseRatingStep
	ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty = metaConsecutivePenalty

	selectionChances := []*config.SelectionChance{
		{MaxThreshold: 0, ChancePercent: 1},
		{MaxThreshold: minRating, ChancePercent: 2},
		{MaxThreshold: maxRating, ChancePercent: 4},
	}

	ratingsConfig.General.SelectionChances = selectionChances

	ratingsData, err := NewRatingsData(ratingsConfig)

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
