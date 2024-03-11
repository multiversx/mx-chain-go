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
		_ = h.Compute("string")
		_ = h.Size()
		_ = h.IsInterfaceNil()
	})
}

func TestHasher_Compute(t *testing.T) {
	t.Parallel()

	h := NewDisabledHasher()
	require.False(t, check.IfNil(h))

	bytes := h.Compute("string")
	require.NotNil(t, bytes)
}

func TestHasher_Size(t *testing.T) {
	t.Parallel()

	h := NewDisabledHasher()
	require.False(t, check.IfNil(h))

	size := h.Size()
	require.NotNil(t, size)
}
