package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppStatusHandler_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	ash := NewAppStatusHandler()
	assert.False(t, check.IfNil(ash))

	require.NotPanics(t, func() {
		ash.AddUint64("key", uint64(0))
		ash.Increment("key")
		ash.Decrement("key")
		ash.SetInt64Value("key", int64(1))
		ash.SetUInt64Value("key", uint64(2))
		ash.SetStringValue("key", "true")
		ash.Close()
	})
}
