package atomic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint32_SetGet(t *testing.T) {
	var number Uint32
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		number.Set(number.Get() + 42)
		wg.Done()
	}()

	go func() {
		number.Set(number.Get() + 43)
		wg.Done()
	}()

	wg.Wait()
	require.Equal(t, uint32(85), number.Get())
}
