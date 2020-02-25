package persister

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetUint64_NotWork(t *testing.T) {
	t.Parallel()

	value := GetUint64(float32(100))

	require.Equal(t, uint64(0), value)
}

func TestGetUint64(t *testing.T) {
	t.Parallel()

	expectedValue := uint64(100)
	value := GetUint64(float64(100))

	require.Equal(t, expectedValue, value)
}
