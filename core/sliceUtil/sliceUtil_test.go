package sliceUtil_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/stretchr/testify/assert"
)

func TestTrimSliceSliceByte_EmptyInputShouldDoNothing(t *testing.T) {
	t.Parallel()

	input := make([][]byte, 0)
	res := sliceUtil.TrimSliceSliceByte(input)

	assert.Equal(t, input, res)
}

func TestTrimSliceSliceByte_ShouldDecreaseCapacity(t *testing.T) {
	t.Parallel()

	input := make([][]byte, 0, 5)
	input = append(input, []byte("el1"))
	input = append(input, []byte("el2"))

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 5, cap(input))

	// after calling the trim func, the capacity should be equal to the len

	input = sliceUtil.TrimSliceSliceByte(input)
	assert.Equal(t, 2, len(input))
	assert.Equal(t, 2, cap(input))
}

func TestTrimSliceSliceByte_SliceAlreadyOkShouldDoNothing(t *testing.T) {
	t.Parallel()

	input := make([][]byte, 0, 2)
	input = append(input, []byte("el1"))
	input = append(input, []byte("el2"))

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 2, cap(input))

	// after calling the trim func, the capacity should be equal to the len

	input = sliceUtil.TrimSliceSliceByte(input)
	assert.Equal(t, 2, len(input))
	assert.Equal(t, 2, cap(input))
}
