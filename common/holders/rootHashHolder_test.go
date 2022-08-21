package holders

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestNewRootHashHolder(t *testing.T) {
	holder := NewRootHashHolder(nil)
	require.Nil(t, holder.GetRootHash())

	holder = NewRootHashHolder([]byte{0xab, 0xcd})
	require.Equal(t, []byte{0xab, 0xcd}, holder.GetRootHash())
	require.Equal(t, core.OptionalUint32{}, holder.GetEpoch())
}

func TestNewRootHashHolderWithEpoch(t *testing.T) {
	holder := NewRootHashHolderWithEpoch([]byte{0xab, 0xcd}, 7)
	require.Equal(t, []byte{0xab, 0xcd}, holder.GetRootHash())
	require.Equal(t, core.OptionalUint32{Value: 7, HasValue: true}, holder.GetEpoch())
}

func TestNewRootHashHolderAsEmpty(t *testing.T) {
	holder := NewRootHashHolderAsEmpty()
	require.Nil(t, holder.GetRootHash())
	require.Equal(t, core.OptionalUint32{}, holder.GetEpoch())
}
