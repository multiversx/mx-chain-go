package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestHasher_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	h := NewDisabledHasher()
	require.False(t, check.IfNil(h))

	require.NotPanics(t, func() {
		bytes := h.Compute("string")
		require.NotNil(t, bytes)

		size := h.Size()
		require.NotNil(t, size)
	})
}
