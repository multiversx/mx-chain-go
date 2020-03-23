package rating

import (
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

	signedBlocksThreshold          = 0.025
	consecutiveMissedBlocksPenalty = 1.1
)

func createDummyRatingsConfig() *config.RatingsConfig {
	return &config.RatingsConfig{
		General: config.General{
			StartRating:           50,
			MaxRating:             100,
			MinRating:             1,
			SignedBlocksThreshold: signedBlocksThreshold,
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

func TestEconomicsData_RatingsDataMinGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsData := createDummyRatingsConfig()
	ratingsData.General.MinRating = 10
	ratingsData.General.MaxRating = 8
	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrMaxRatingIsSmallerThanMinRating, err)
}

func TestEconomicsData_RatingsDataMinSmallerThanOne(t *testing.T) {
	t.Parallel()

	ratingsData := createDummyRatingsConfig()
	ratingsData.General.MinRating = 0
	ratingsData.General.MaxRating = 8
	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrMinRatingSmallerThanOne, err)
}

func TestEconomicsData_RatingsStartGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	ratingsData := createDummyRatingsConfig()
	ratingsData.General.MinRating = 10
	ratingsData.General.MaxRating = 100
	ratingsData.General.StartRating = 110
	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrStartRatingNotBetweenMinAndMax, err)
}

func TestEconomicsData_RatingsStartLowerMinShouldErr(t *testing.T) {
	t.Parallel()

	ratingsData := createDummyRatingsConfig()
	ratingsData.General.MinRating = 10
	ratingsData.General.MaxRating = 100
	ratingsData.General.StartRating = 5
	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrStartRatingNotBetweenMinAndMax, err)
}

func TestEconomicsData_RatingsSignedBlocksThresholdNotBetweenZeroAndOneShouldErr(t *testing.T) {
	t.Parallel()

	ratingsData := createDummyRatingsConfig()
	ratingsData.General.SignedBlocksThreshold = -0.1
	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne, err)

	ratingsData.General.SignedBlocksThreshold = 1.01
	economicsData, err = NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne, err)
}

func TestEconomicsData_RatingsConsecutiveMissedBlocksPenaltyLowerThanOneShouldErr(t *testing.T) {
	t.Parallel()

	ratingsData := createDummyRatingsConfig()
	ratingsData.MetaChain.ConsecutiveMissedBlocksPenalty = 0.9
	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne, err)

	ratingsData.ShardChain.ConsecutiveMissedBlocksPenalty = 0.99
	economicsData, err = NewRatingsData(ratingsData)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne, err)
}

func TestEconomicsData_RatingsCorrectValues(t *testing.T) {
	t.Parallel()

	minRating := uint32(10)
	maxRating := uint32(100)
	startRating := uint32(50)
	signedBlocksThreshold := float32(0.025)
	shardConsecutivePenalty := float32(1.2)
	metaConsecutivePenalty := float32(1.3)
	ratingsData := createDummyRatingsConfig()
	ratingsData.General.MinRating = minRating
	ratingsData.General.MaxRating = maxRating
	ratingsData.General.StartRating = startRating
	ratingsData.General.SignedBlocksThreshold = signedBlocksThreshold
	ratingsData.ShardChain.ProposerDecreaseRatingStep = shardProposerDecreaseRatingStep
	ratingsData.ShardChain.ProposerIncreaseRatingStep = shardProposerIncreaseRatingStep
	ratingsData.ShardChain.ValidatorIncreaseRatingStep = shardValidatorIncreaseRatingStep
	ratingsData.ShardChain.ValidatorDecreaseRatingStep = shardValidatorDecreaseRatingStep
	ratingsData.ShardChain.ConsecutiveMissedBlocksPenalty = shardConsecutivePenalty
	ratingsData.MetaChain.ProposerDecreaseRatingStep = metaProposerDecreaseRatingStep
	ratingsData.MetaChain.ProposerIncreaseRatingStep = metaProposerIncreaseRatingStep
	ratingsData.MetaChain.ValidatorIncreaseRatingStep = metaValidatorIncreaseRatingStep
	ratingsData.MetaChain.ValidatorDecreaseRatingStep = metaValidatorDecreaseRatingStep
	ratingsData.MetaChain.ConsecutiveMissedBlocksPenalty = metaConsecutivePenalty

	economicsData, err := NewRatingsData(ratingsData)

	assert.Nil(t, err)
	assert.NotNil(t, economicsData)
	assert.Equal(t, startRating, economicsData.StartRating())
	assert.Equal(t, minRating, economicsData.MinRating())
	assert.Equal(t, maxRating, economicsData.MaxRating())
	assert.Equal(t, signedBlocksThreshold, economicsData.SignedBlocksThreshold())
	assert.Equal(t, shardValidatorIncreaseRatingStep, economicsData.ShardChainRatingsStepHandler().ValidatorIncreaseRatingStep())
	assert.Equal(t, shardValidatorDecreaseRatingStep, economicsData.ShardChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	assert.Equal(t, shardProposerIncreaseRatingStep, economicsData.ShardChainRatingsStepHandler().ProposerIncreaseRatingStep())
	assert.Equal(t, shardProposerDecreaseRatingStep, economicsData.ShardChainRatingsStepHandler().ProposerDecreaseRatingStep())
	assert.Equal(t, metaValidatorIncreaseRatingStep, economicsData.MetaChainRatingsStepHandler().ValidatorIncreaseRatingStep())
	assert.Equal(t, metaValidatorDecreaseRatingStep, economicsData.MetaChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	assert.Equal(t, metaProposerIncreaseRatingStep, economicsData.MetaChainRatingsStepHandler().ProposerIncreaseRatingStep())
	assert.Equal(t, metaProposerDecreaseRatingStep, economicsData.MetaChainRatingsStepHandler().ProposerDecreaseRatingStep())
	assert.Equal(t, shardConsecutivePenalty, economicsData.ShardChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty())
	assert.Equal(t, metaConsecutivePenalty, economicsData.MetaChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty())
}
