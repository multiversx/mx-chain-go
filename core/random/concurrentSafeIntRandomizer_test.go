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
				_, err := csir.Intn(idx + 1)
				assert.Nil(t, err)

				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	fmt.Println("Waiting 10 seconds...")
	time.Sleep(time.Second * 10)
}

func TestConcurrentSafeIntRandomizer_IntnInvalidParamShouldPanic(t *testing.T) {
	defer func() {
		r := recover()

		assert.NotNil(t, r)
	}()
	csir := &ConcurrentSafeIntRandomizer{}
	assert.False(t, csir.IsInterfaceNil())

	res, err := csir.Intn(-1)
	assert.Equal(t, 0, res)
	assert.NotNil(t, err)
}
