package random

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentSafeIntRandomizer_IntnConcurrent(t *testing.T) {
	csir := &ConcurrentSafeIntRandomizer{}

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("randomizer fail: %v", t))
		}
	}()

	maxIterations := 100
	for i := 0; i < maxIterations; i++ {
		go func(idx int) {
			for {
				_ = csir.Intn(idx + 1)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	fmt.Println("Waiting 1 second...")
	time.Sleep(time.Second * 1)
}

func TestConcurrentSafeIntRandomizer_IntnInvalidShouldReturnZero(t *testing.T) {
	t.Parallel()

	csir := &ConcurrentSafeIntRandomizer{}
	assert.False(t, csir.IsInterfaceNil())

	res := csir.Intn(-1)
	assert.Equal(t, 0, res)

	res = csir.Intn(0)
	assert.Equal(t, 0, res)

	res = csir.Intn(1)
	assert.Equal(t, 0, res)
}

func TestConcurrentSafeIntRandomizer_IntnShouldWork(t *testing.T) {
	t.Parallel()

	csir := &ConcurrentSafeIntRandomizer{}
	assert.False(t, csir.IsInterfaceNil())

	maxValue := 70
	res := csir.Intn(maxValue)

	assert.True(t, res >= 0, fmt.Sprintf("0 comparison, generated %d, max %d", res, maxValue))
	assert.True(t, res < maxValue, fmt.Sprintf("max comparison, generated %d, max %d", res, maxValue))
}
