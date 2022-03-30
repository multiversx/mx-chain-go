package leveldb

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/stretchr/testify/assert"
)

func TestWaitingOperations_IncrementDecrementGetCounter(t *testing.T) {
	t.Parallel()

	pattern := "%s: %d"

	counters := newWaitingOperationsCounters()
	assert.Equal(t, 0, counters.getCounter(common.LowPriority))
	assert.Equal(t, 0, counters.getCounter(common.HighPriority))
	assert.Equal(t, "", counters.snapshotString())

	counters.increment(common.LowPriority)
	assert.Equal(t, 1, counters.getCounter(common.LowPriority))
	assert.Equal(t, 0, counters.getCounter(common.HighPriority))
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 1), counters.snapshotString())

	counters.increment(common.LowPriority)
	assert.Equal(t, 2, counters.getCounter(common.LowPriority))
	assert.Equal(t, 0, counters.getCounter(common.HighPriority))
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 2), counters.snapshotString())

	counters.decrement(common.LowPriority)
	assert.Equal(t, 1, counters.getCounter(common.LowPriority))
	assert.Equal(t, 0, counters.getCounter(common.HighPriority))
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 1), counters.snapshotString())

	counters.decrement(common.LowPriority)
	assert.Equal(t, 0, counters.getCounter(common.LowPriority))
	assert.Equal(t, 0, counters.getCounter(common.HighPriority))
	assert.Equal(t, fmt.Sprintf(pattern, common.LowPriority, 0), counters.snapshotString())

	counters.increment(common.LowPriority)
	counters.increment(common.HighPriority)
	assert.Equal(t, 1, counters.getCounter(common.LowPriority))
	assert.Equal(t, 1, counters.getCounter(common.HighPriority))
	expectedString := strings.Join([]string{
		fmt.Sprintf(pattern, common.HighPriority, 1),
		fmt.Sprintf(pattern, common.LowPriority, 1),
	}, ", ")

	assert.Equal(t, expectedString, counters.snapshotString())
}

func TestWaitingOperations_ParallelOperations(t *testing.T) {
	t.Parallel()

	numOperations := 120
	counters := newWaitingOperationsCounters()
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		local := i % 4
		go func() {
			switch local {
			case 0:
				_ = counters.getCounter(common.LowPriority)
			case 1:
				counters.increment(common.LowPriority)
			case 2:
				counters.decrement(common.LowPriority)
			case 3:
				_ = counters.snapshotString()
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
