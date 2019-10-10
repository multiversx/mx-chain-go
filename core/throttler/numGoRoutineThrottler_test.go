package throttler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/stretchr/testify/assert"
)

func TestNewNumGoRoutineThrottler_WithNegativeShouldError(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumGoRoutineThrottler(-1)

	assert.Nil(t, nt)
	assert.Equal(t, core.ErrNotPositiveValue, err)
}

func TestNewNumGoRoutineThrottler_WithZeroShouldError(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumGoRoutineThrottler(0)

	assert.Nil(t, nt)
	assert.Equal(t, core.ErrNotPositiveValue, err)
}

func TestNewNumGoRoutineThrottler_ShouldWork(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumGoRoutineThrottler(1)

	assert.NotNil(t, nt)
	assert.Nil(t, err)
}

func TestNumGoRoutineThrottler_CanProcessMessageWithZeroCounter(t *testing.T) {
	t.Parallel()

	nt, _ := throttler.NewNumGoRoutineThrottler(1)

	assert.True(t, nt.CanProcess())
}

func TestNumGoRoutineThrottler_CanProcessMessageCounterEqualsMax(t *testing.T) {
	t.Parallel()

	nt, _ := throttler.NewNumGoRoutineThrottler(1)
	nt.StartProcessing()

	assert.False(t, nt.CanProcess())
}

func TestNumGoRoutineThrottler_CanProcessMessageCounterIsMaxLessThanOne(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumGoRoutineThrottler(max)

	for i := int32(0); i < max-1; i++ {
		nt.StartProcessing()
	}

	assert.True(t, nt.CanProcess())
}

func TestNumGoRoutineThrottler_CanProcessMessageCounterIsMax(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumGoRoutineThrottler(max)

	for i := int32(0); i < max; i++ {
		nt.StartProcessing()
	}

	assert.False(t, nt.CanProcess())
}

func TestNumGoRoutineThrottler_CanProcessMessageCounterIsMaxLessOneFromEndProcessMessage(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumGoRoutineThrottler(max)

	for i := int32(0); i < max; i++ {
		nt.StartProcessing()
	}
	nt.EndProcessing()

	assert.True(t, nt.CanProcess())
}
