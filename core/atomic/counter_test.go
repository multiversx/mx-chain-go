package atomic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounter_IncrementAndDecrement(t *testing.T) {
	var counter Counter
	var wg sync.WaitGroup

	// Increment 100 * 100 times
	// Decrement 100 * 50 times
	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func() {
			for j := 0; j < 100; j++ {
				counter.Increment()
			}

			wg.Done()
		}()

		go func() {
			for j := 0; j < 50; j++ {
				counter.Decrement()
			}

			wg.Done()
		}()
	}

	wg.Wait()

	require.Equal(t, int64(5000), counter.Get())
	require.Equal(t, uint64(5000), counter.GetUint64())
}

func TestCounter_AddAndSubtract(t *testing.T) {
	var counter Counter
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		for j := 0; j < 10; j++ {
			counter.Add(5)
		}

		wg.Done()
	}()

	go func() {
		for j := 0; j < 10; j++ {
			counter.Subtract(4)
		}

		wg.Done()
	}()

	wg.Wait()

	require.Equal(t, int64(10), counter.Get())
	require.Equal(t, uint64(10), counter.GetUint64())
}

func TestCounter_SetAndReset(t *testing.T) {
	var counter Counter
	var wg sync.WaitGroup

	counter.Set(41)

	wg.Add(2)

	go func() {
		counter.Set(42)
		wg.Done()
	}()

	go func() {
		counter.Reset()
		wg.Done()
	}()

	wg.Wait()

	require.True(t, counter.Get() == 0 || counter.Get() == 42)
}

func TestCounter_GetUint64(t *testing.T) {
	var counter Counter

	counter.Set(12345)
	require.Equal(t, int64(12345), counter.Get())

	counter.Set(-42)
	require.Equal(t, uint64(0), counter.GetUint64())
}
