package atomic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestString_SetGet(t *testing.T) {
	var str String
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		str.Set(str.Get() + "foo")
		wg.Done()
	}()

	go func() {
		str.Set(str.Get() + "bar")
		wg.Done()
	}()

	wg.Wait()
	require.Contains(t, []string{"foobar", "barfoo", "foo", "bar"}, str.Get())
}
