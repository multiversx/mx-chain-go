package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderedCollection_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	oc := NewOrderedCollection()
	assert.False(t, check.IfNil(oc))

	require.NotPanics(t, func() {
		oc.Add([]byte("key"))
		_, _ = oc.GetItemAtIndex(0)
		_, _ = oc.GetOrder([]byte("key"))
		oc.Remove([]byte("key"))
		oc.RemoveMultiple([][]byte{[]byte("key1"), []byte("key2")})
		_ = oc.GetItems()
		_ = oc.Contains([]byte("key"))
		oc.Clear()
		_ = oc.Len()
		_ = oc.IsInterfaceNil()
	})
}
