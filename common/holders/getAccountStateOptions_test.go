package holders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGetAccountStateOptions(t *testing.T) {
	options := NewGetAccountStateOptions(nil)
	require.Nil(t, options.GetBlockRootHash())

	options = NewGetAccountStateOptions([]byte{0xab, 0xcd})
	require.Equal(t, []byte{0xab, 0xcd}, options.GetBlockRootHash())
}

func TestNewGetAccountStateOptionsAsEmpty(t *testing.T) {
	options := NewGetAccountStateOptionsAsEmpty()
	require.Nil(t, options.GetBlockRootHash())
}
