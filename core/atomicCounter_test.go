package core_test

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestIncrementAndDecrement(t *testing.T) {
	var counter core.AtomicCounter
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
