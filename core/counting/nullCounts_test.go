package counting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullCounts(t *testing.T) {
	counts := NullCounts{}

	total := counts.GetTotal()
	asString := counts.String()

	require.Equal(t, int64(-1), total)
	require.Equal(t, asString, "counts not applicable")
}

func TestNullCounts_IsInterfaceNil(t *testing.T) {
	counts := &NullCounts{}
	require.False(t, counts.IsInterfaceNil())

	thisIsNil := (*NullCounts)(nil)
	require.True(t, thisIsNil.IsInterfaceNil())
}
