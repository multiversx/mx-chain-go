package display_test

import (
	"encoding/hex"
	"strings"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/stretchr/testify/assert"
)

func TestSetDisplayByteSlice_NilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	err := display.SetDisplayByteSlice(nil)
	assert.Equal(t, display.ErrNilDisplayByteSliceHandler, err)
}

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

func TestToHexShort_EmptySliceShouldReturnEmptyString(t *testing.T) {
	t.Parallel()

	hash := []byte("")
	res := display.ToHexShort(hash)

	assert.Equal(t, 0, len(res))
}

func TestToHexShort_SliceLengthSmallShouldNotChange(t *testing.T) {
	t.Parallel()

	input := []byte("short")
	hexHash := hex.EncodeToString(input)

	res := display.ToHexShort(input)

	assert.Equal(t, hexHash, res)
}

func TestToHexShort_SliceLengthBigShouldTrim(t *testing.T) {
	t.Parallel()

	input := []byte("long enough input so it should be trimmed")
	hexHash := hex.EncodeToString(input)

	res := display.ToHexShort(input)

	assert.NotEqual(t, len(hexHash), len(res))
	assert.True(t, strings.Contains(res, "..."))
}
