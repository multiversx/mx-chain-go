package persister

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetUint64(t *testing.T) {
	t.Parallel()

	expectedValue := uint64(100)
	value := GetUint64(uint64(100))

	require.Equal(t, expectedValue, value)
}

func TestGetString(t *testing.T) {
	t.Parallel()

	expectedValue := "tesstt"
	value := GetString("tesstt")

	require.Equal(t, expectedValue, value)
}
