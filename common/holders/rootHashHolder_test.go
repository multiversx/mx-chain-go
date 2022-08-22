package holders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRootHashHolder(t *testing.T) {
	holder := NewRootHashHolder(nil)
	require.Nil(t, holder.GetRootHash())

	holder = NewRootHashHolder([]byte{0xab, 0xcd})
	require.Equal(t, []byte{0xab, 0xcd}, holder.GetRootHash())
}

func TestNewRootHashHolderAsEmpty(t *testing.T) {
	holder := NewRootHashHolderAsEmpty()
	require.Nil(t, holder.GetRootHash())
}
