package display_test

import (
	"encoding/hex"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/stretchr/testify/assert"
)

func TestDisplayByteSlice_WithSetCanBeCalledConcurrently(t *testing.T) {
	t.Parallel()

	numCalls := 100
	wg := &sync.WaitGroup{}
	wg.Add(2 * numCalls)
	for i := 0; i < numCalls; i++ {
		go func() {
			_ = display.DisplayByteSlice([]byte("sss"))

			wg.Done()
		}()

		go func() {
			err := display.SetDisplayByteSlice(func(slice []byte) string {
				return hex.EncodeToString(slice)
			})

			assert.Nil(t, err)

			wg.Done()
		}()
	}

	wg.Wait()
}
