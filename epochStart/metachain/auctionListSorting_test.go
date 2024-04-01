package metachain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalcNormalizedRandomness(t *testing.T) {
	t.Parallel()

	t.Run("randomness longer than expected len", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 2)
		require.Equal(t, []byte("ra"), result)
	})

	t.Run("randomness length equal to expected len", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 4)
		require.Equal(t, []byte("rand"), result)
	})

	t.Run("randomness length less than expected len", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 6)
		require.Equal(t, []byte("randra"), result)
	})

	t.Run("expected len is zero", func(t *testing.T) {
		t.Parallel()

		result := calcNormalizedRandomness([]byte("rand"), 0)
		require.Empty(t, result)
	})
}
