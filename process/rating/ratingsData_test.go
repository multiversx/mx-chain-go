package rating

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

const (
	shardValidatorIncreaseRatingStep = int32(2)
	shardValidatorDecreaseRatingStep = int32(4)
	shardProposerIncreaseRatingStep  = int32(1)
	shardProposerDecreaseRatingStep  = int32(2)

	metaValidatorIncreaseRatingStep = int32(2)
	metaValidatorDecreaseRatingStep = int32(4)
	metaProposerIncreaseRatingStep  = int32(1)
	metaProposerDecreaseRatingStep  = int32(2)
)

func createDummyRatingsConfig() config.RatingsConfig {
	return config.RatingsConfig{
		General: config.General{
			StartRating: 50,
			MaxRating:   100,
			MinRating:   1,
		},
		ShardChain: config.ShardChain{
			RatingSteps: config.RatingSteps{
				ProposerIncreaseRatingStep:  shardProposerIncreaseRatingStep,
				ProposerDecreaseRatingStep:  shardProposerDecreaseRatingStep,
				ValidatorIncreaseRatingStep: shardValidatorIncreaseRatingStep,
				ValidatorDecreaseRatingStep: shardValidatorDecreaseRatingStep,
			},
		},
		MetaChain: config.MetaChain{
			RatingSteps: config.RatingSteps{
				ProposerIncreaseRatingStep:  metaProposerIncreaseRatingStep,
				ProposerDecreaseRatingStep:  metaProposerDecreaseRatingStep,
				ValidatorIncreaseRatingStep: metaValidatorIncreaseRatingStep,
				ValidatorDecreaseRatingStep: metaValidatorDecreaseRatingStep,
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

func TestRatingsData_RatingsCorrectValues(t *testing.T) {
	t.Parallel()

	minRating := uint32(10)
	maxRating := uint32(100)
	startRating := uint32(50)

	ratingsConfig := createDummyRatingsConfig()
	ratingsConfig.General.MinRating = minRating
	ratingsConfig.General.MaxRating = maxRating
	ratingsConfig.General.StartRating = startRating
	ratingsConfig.ShardChain.ProposerDecreaseRatingStep = shardProposerDecreaseRatingStep
	ratingsConfig.ShardChain.ProposerIncreaseRatingStep = shardProposerIncreaseRatingStep
	ratingsConfig.ShardChain.ValidatorIncreaseRatingStep = shardValidatorIncreaseRatingStep
	ratingsConfig.ShardChain.ValidatorDecreaseRatingStep = shardValidatorDecreaseRatingStep
	ratingsConfig.MetaChain.ProposerDecreaseRatingStep = metaProposerDecreaseRatingStep
	ratingsConfig.MetaChain.ProposerIncreaseRatingStep = metaProposerIncreaseRatingStep
	ratingsConfig.MetaChain.ValidatorIncreaseRatingStep = metaValidatorIncreaseRatingStep
	ratingsConfig.MetaChain.ValidatorDecreaseRatingStep = metaValidatorDecreaseRatingStep

	ratingsData, err := NewRatingsData(ratingsConfig)

	assert.Nil(t, err)
	assert.NotNil(t, ratingsData)
	assert.Equal(t, startRating, ratingsData.StartRating())
	assert.Equal(t, minRating, ratingsData.MinRating())
	assert.Equal(t, maxRating, ratingsData.MaxRating())
	assert.Equal(t, shardValidatorIncreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ValidatorIncreaseRatingStep())
	assert.Equal(t, shardValidatorDecreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	assert.Equal(t, shardProposerIncreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ProposerIncreaseRatingStep())
	assert.Equal(t, shardProposerDecreaseRatingStep, ratingsData.ShardChainRatingsStepHandler().ProposerDecreaseRatingStep())
	assert.Equal(t, metaValidatorIncreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ValidatorIncreaseRatingStep())
	assert.Equal(t, metaValidatorDecreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	assert.Equal(t, metaProposerIncreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ProposerIncreaseRatingStep())
	assert.Equal(t, metaProposerDecreaseRatingStep, ratingsData.MetaChainRatingsStepHandler().ProposerDecreaseRatingStep())
}
