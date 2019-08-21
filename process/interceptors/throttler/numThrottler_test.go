package throttler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/throttler"
	"github.com/stretchr/testify/assert"
)

func TestNewNumThrottler_WithNegativeShouldError(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumThrottler(-1)

	assert.Nil(t, nt)
	assert.Equal(t, process.ErrNotPositiveValue, err)
}

func TestNewNumThrottler_WithZeroShouldError(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumThrottler(0)

	assert.Nil(t, nt)
	assert.Equal(t, process.ErrNotPositiveValue, err)
}

func TestNewNumThrottler_ShouldWork(t *testing.T) {
	t.Parallel()

	nt, err := throttler.NewNumThrottler(1)

	assert.NotNil(t, nt)
	assert.Nil(t, err)
}

func TestNumThrottler_CanProcessMessageWithZeroCounter(t *testing.T) {
	t.Parallel()

	nt, _ := throttler.NewNumThrottler(1)

	assert.True(t, nt.CanProcess())
}

func TestNumThrottler_CanProcessMessageCounterEqualsMax(t *testing.T) {
	t.Parallel()

	nt, _ := throttler.NewNumThrottler(1)
	nt.StartProcessing()

	assert.False(t, nt.CanProcess())
}

func TestNumThrottler_CanProcessMessageCounterIsMaxLessThanOne(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumThrottler(max)

	for i := int32(0); i < max-1; i++ {
		nt.StartProcessing()
	}

	assert.True(t, nt.CanProcess())
}

func TestNumThrottler_CanProcessMessageCounterIsMax(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumThrottler(max)

	for i := int32(0); i < max; i++ {
		nt.StartProcessing()
	}

	assert.False(t, nt.CanProcess())
}

func TestNumThrottler_CanProcessMessageCounterIsMaxLessOneFromEndProcessMessage(t *testing.T) {
	t.Parallel()

	max := int32(45)
	nt, _ := throttler.NewNumThrottler(max)

	for i := int32(0); i < max; i++ {
		nt.StartProcessing()
	}
	nt.EndProcessing()

	assert.True(t, nt.CanProcess())
}
