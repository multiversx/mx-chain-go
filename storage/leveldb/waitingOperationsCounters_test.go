package leveldb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/stretchr/testify/assert"
)

func TestWaitingOperations_IncrementDecrementGetCounter(t *testing.T) {
	t.Parallel()

	pattern := "%s: %d"

	counters := newWaitingOperationsCounters()
	assert.Equal(t, 0, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.counters[common.HighPriority])
	assert.Equal(t, "", counters.snapshotString())
	assert.Equal(t, 0, counters.numNonZeroCounters)

	counters.increment(common.LowPriority)
	assert.Equal(t, 1, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.counters[common.HighPriority])
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 1), counters.snapshotString())
	assert.Equal(t, 1, counters.numNonZeroCounters)

	counters.increment(common.LowPriority)
	assert.Equal(t, 2, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.counters[common.HighPriority])
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 2), counters.snapshotString())
	assert.Equal(t, 1, counters.numNonZeroCounters)

	counters.decrement(common.LowPriority)
	assert.Equal(t, 1, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.counters[common.HighPriority])
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 1), counters.snapshotString())
	assert.Equal(t, 1, counters.numNonZeroCounters)

	counters.decrement(common.LowPriority)
	assert.Equal(t, 0, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.counters[common.HighPriority])
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 0), counters.snapshotString())
	assert.Equal(t, 0, counters.numNonZeroCounters)

	counters.increment(common.LowPriority)
	counters.increment(common.HighPriority)
	assert.Equal(t, 1, counters.counters[common.LowPriority])
	assert.Equal(t, 1, counters.counters[common.HighPriority])
	expectedString := strings.Join([]string{
		fmt.Sprintf(pattern, common.HighPriority, 1),
		fmt.Sprintf(pattern, common.LowPriority, 1),
	}, ", ")
	assert.Equal(t, 2, counters.numNonZeroCounters)

	assert.Equal(t, expectedString, counters.snapshotString())

	counters.decrement(common.LowPriority)
	assert.Equal(t, 1, counters.numNonZeroCounters)
	counters.decrement(common.HighPriority)
	assert.Equal(t, 0, counters.numNonZeroCounters)
}

func TestWaitingOperations_decrementUnderZero(t *testing.T) {
	t.Parallel()

	counters := newWaitingOperationsCounters()
	assert.Equal(t, 0, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.numNonZeroCounters)

	counters.decrement(common.LowPriority)
	assert.Equal(t, 0, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.numNonZeroCounters)

	counters.increment(common.LowPriority)
	assert.Equal(t, 1, counters.counters[common.LowPriority])
	assert.Equal(t, 1, counters.numNonZeroCounters)

	counters.increment(common.LowPriority)
	assert.Equal(t, 2, counters.counters[common.LowPriority])
	assert.Equal(t, 1, counters.numNonZeroCounters)

	counters.decrement(common.LowPriority)
	counters.decrement(common.LowPriority)
	counters.decrement(common.LowPriority)
	counters.decrement(common.LowPriority)

	counters.decrement(common.LowPriority)
	assert.Equal(t, 0, counters.counters[common.LowPriority])
	assert.Equal(t, 0, counters.numNonZeroCounters)
}
