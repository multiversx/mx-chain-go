package holders

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestNewRootHashHolder(t *testing.T) {
	holder := NewRootHashHolder(nil, core.OptionalUint32{})
	require.Nil(t, holder.GetRootHash())
	require.Equal(t, core.OptionalUint32{}, holder.GetEpoch())

	holder = NewRootHashHolder([]byte{0xab, 0xcd}, core.OptionalUint32{Value: 7, HasValue: true})
	require.Equal(t, []byte{0xab, 0xcd}, holder.GetRootHash())
	require.Equal(t, core.OptionalUint32{Value: 7, HasValue: true}, holder.GetEpoch())
}

func TestNewRootHashHolderAsEmpty(t *testing.T) {
	holder := NewRootHashHolderAsEmpty()
	require.Nil(t, holder.GetRootHash())
	require.Equal(t, core.OptionalUint32{}, holder.GetEpoch())
}
