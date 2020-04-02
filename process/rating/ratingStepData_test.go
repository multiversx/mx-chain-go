package rating

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
)

func TestRatingStepData_NewRatingStepDataShouldWork(t *testing.T) {
	t.Parallel()

	proposerIncreaseRatingStep := int32(1)
	proposerDecreaseRatingStep := int32(-2)
	validatorIncreaseRatingStep := int32(3)
	validatorDecreaseRatingStep := int32(-4)

	steps := config.RatingSteps{
		ProposerIncreaseRatingStep:  proposerIncreaseRatingStep,
		ProposerDecreaseRatingStep:  proposerDecreaseRatingStep,
		ValidatorIncreaseRatingStep: validatorIncreaseRatingStep,
		ValidatorDecreaseRatingStep: validatorDecreaseRatingStep,
	}

	rsd := NewRatingStepData(steps)

	assert.NotNil(t, rsd)
	assert.Equal(t, proposerIncreaseRatingStep, rsd.ProposerIncreaseRatingStep())
	assert.Equal(t, proposerDecreaseRatingStep, rsd.ProposerDecreaseRatingStep())
	assert.Equal(t, validatorIncreaseRatingStep, rsd.ValidatorIncreaseRatingStep())
	assert.Equal(t, validatorDecreaseRatingStep, rsd.ValidatorDecreaseRatingStep())
}
