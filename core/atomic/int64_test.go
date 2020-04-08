package atomic

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt64_SetGet(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have panicked %v", r))
		}
	}()

	var number Int64
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		number.Set(number.Get() - 42)
		wg.Done()
	}()

	go func() {
		number.Set(number.Get() + 43)
		wg.Done()
	}()

	wg.Wait()
}
