package throttler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/stretchr/testify/assert"
)

func TestNewNumGoRoutinesThrottler_WithNegativeShouldError(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumGoRoutinesThrottler(-1)

	assert.Nil(t, nt)
	assert.Equal(t, core.ErrNotPositiveValue, err)
}

func TestNewNumGoRoutinesThrottler_WithZeroShouldError(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumGoRoutinesThrottler(0)

	assert.Nil(t, nt)
	assert.Equal(t, core.ErrNotPositiveValue, err)
}

func TestNewNumGoRoutinesThrottler_ShouldWork(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumGoRoutinesThrottler(1)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(nt))
}

func TestNumGoRoutinesThrottler_CanProcessMessageWithZeroCounter(t *testing.T) {
	t.Parallel()

	nt, _ := throttler.NewNumGoRoutinesThrottler(1)

	assert.True(t, nt.CanProcess())
}

func TestNumGoRoutinesThrottler_CanProcessMessageCounterEqualsMax(t *testing.T) {
	t.Parallel()

	nt, _ := throttler.NewNumGoRoutinesThrottler(1)
	nt.StartProcessing()

	assert.False(t, nt.CanProcess())
}

func TestNumGoRoutinesThrottler_CanProcessMessageCounterIsMaxLessThanOne(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumGoRoutinesThrottler(max)

	for i := int32(0); i < max-1; i++ {
		nt.StartProcessing()
	}

	assert.True(t, nt.CanProcess())
}

func TestNumGoRoutinesThrottler_CanProcessMessageCounterIsMax(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumGoRoutinesThrottler(max)

	for i := int32(0); i < max; i++ {
		nt.StartProcessing()
	}

	assert.False(t, nt.CanProcess())
}

func TestNumGoRoutinesThrottler_CanProcessMessageCounterIsMaxLessOneFromEndProcessMessage(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumGoRoutinesThrottler(max)

	for i := int32(0); i < max; i++ {
		nt.StartProcessing()
	}
	nt.EndProcessing()

	assert.True(t, nt.CanProcess())
}
