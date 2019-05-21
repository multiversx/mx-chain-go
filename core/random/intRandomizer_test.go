package random

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntRandomizerConcurrentSafe_IntnConcurrent(t *testing.T) {
	ircs := &IntRandomizerConcurrentSafe{}

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("randomizer fail: %v", t))
		}
	}()

	for i := 0; i < 1000; i++ {
		go func(idx int) {
			for {
				_, err := ircs.Intn(idx + 1)
				assert.Nil(t, err)

				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	fmt.Println("Waiting 10 seconds...")
	time.Sleep(time.Second * 10)
}
