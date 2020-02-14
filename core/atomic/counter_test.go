package atomic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, int64(5000), counter.Get())
}

func TestCounter_SetAndGet(t *testing.T) {
	t.Parallel()

	var counter Counter

	value := int64(10)
	counter.Set(value)

	assert.Equal(t, value, counter.Get())
}
