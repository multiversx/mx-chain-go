package statistics

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieSyncStatistics_ShouldWork(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.False(t, check.IfNil(tss))
}

func TestTrieSyncStatistics_Processed(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumProcessed())

	tss.AddNumProcessed(2)
	assert.Equal(t, 2, tss.NumProcessed())

	tss.AddNumProcessed(4)
	assert.Equal(t, 6, tss.NumProcessed())

	tss.Reset()
	assert.Equal(t, 0, tss.NumProcessed())
}

func TestTrieSyncStatistics_Missing(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumMissing())
	assert.Equal(t, 0, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 2)
	assert.Equal(t, 2, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 4)
	assert.Equal(t, 4, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.SetNumMissing([]byte("rh2"), 6)
	assert.Equal(t, 10, tss.NumMissing())
	assert.Equal(t, 2, tss.NumTries())

	tss.SetNumMissing([]byte("rh3"), 0)
	assert.Equal(t, 10, tss.NumMissing())
	assert.Equal(t, 2, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 0)
	assert.Equal(t, 6, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.SetNumMissing([]byte("rh2"), 0)
	assert.Equal(t, 0, tss.NumMissing())
	assert.Equal(t, 0, tss.NumTries())

	tss.SetNumMissing([]byte("rh1"), 67)
	assert.Equal(t, 67, tss.NumMissing())
	assert.Equal(t, 1, tss.NumTries())

	tss.Reset()
	assert.Equal(t, 0, tss.NumMissing())
	assert.Equal(t, 0, tss.NumTries())
}

func TestTrieSyncStatistics_Large(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumLarge())

	tss.AddNumLarge(2)
	assert.Equal(t, 2, tss.NumLarge())

	tss.AddNumLarge(4)
	assert.Equal(t, 6, tss.NumLarge())

	tss.Reset()
	assert.Equal(t, 0, tss.NumLarge())
}

func TestTrieSyncStatistics_BytesReceived(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, uint64(0), tss.NumBytesReceived())

	tss.AddNumBytesReceived(2)
	assert.Equal(t, uint64(2), tss.NumBytesReceived())

	tss.AddNumBytesReceived(4)
	assert.Equal(t, uint64(6), tss.NumBytesReceived())

	tss.Reset()
	assert.Equal(t, uint64(0), tss.NumBytesReceived())
}

func TestTrieSyncStatistics_IncrementIteration(t *testing.T) {
	t.Parallel()

	tss := NewTrieSyncStatistics()

	assert.Equal(t, 0, tss.NumIterations())

	tss.IncrementIteration()
	assert.Equal(t, 1, tss.NumIterations())

	tss.IncrementIteration()
	assert.Equal(t, 2, tss.NumIterations())

	tss.Reset()
	assert.Equal(t, 0, tss.NumIterations())
}

func TestTrieSyncStatistics_AddProcessingTime(t *testing.T) {
	t.Parallel()

	t.Run("one go routine", func(t *testing.T) {
		tss := NewTrieSyncStatistics()

		assert.Equal(t, time.Duration(0), tss.ProcessingTime())

		tss.AddProcessingTime(time.Second)
		assert.Equal(t, time.Second, tss.ProcessingTime())

		tss.AddProcessingTime(time.Millisecond)
		assert.Equal(t, time.Second+time.Millisecond, tss.ProcessingTime())

		tss.Reset()
		assert.Equal(t, time.Duration(0), tss.ProcessingTime())
	})
	t.Run("more go routines", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		tss := NewTrieSyncStatistics()

		numGoRoutines := 10
		processingTime := time.Millisecond * 50
		wg.Add(numGoRoutines)
		for i := 0; i < numGoRoutines; i++ {
			go func() {
				time.Sleep(time.Millisecond * 10)

				tss.AddProcessingTime(processingTime)
				wg.Done()
			}()
		}

		wg.Wait()

		assert.Equal(t, time.Duration(numGoRoutines)*processingTime, tss.ProcessingTime())
	})
}

func TestParallelOperationsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	tss := NewTrieSyncStatistics()
	numIterations := 10000
	wg := sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			switch idx {
			case 0:
				tss.Reset()
			case 1:
				tss.AddNumProcessed(1)
			case 2:
				tss.AddNumBytesReceived(2)
			case 3:
				tss.AddNumLarge(3)
			case 4:
				tss.SetNumMissing([]byte("root hash"), 4)
			case 5:
				tss.AddProcessingTime(time.Millisecond)
			case 6:
				tss.IncrementIteration()
			case 7:
				_ = tss.NumProcessed()
			case 8:
				_ = tss.NumLarge()
			case 9:
				_ = tss.NumMissing()
			case 10:
				_ = tss.NumBytesReceived()
			case 11:
				_ = tss.NumTries()
			case 12:
				_ = tss.ProcessingTime()
			case 13:
				_ = tss.NumIterations()
			}

			wg.Done()
		}(i % 14)
	}

	wg.Wait()
}
