package holders

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
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

func TestNewRootHashHolder_String(t *testing.T) {
	holder := NewRootHashHolder(
		[]byte("rootHash"),
		core.OptionalUint32{
			Value:    5,
			HasValue: true,
		},
	)
	hexRootHash := hex.EncodeToString([]byte("rootHash"))
	expectedString := "root hash " + hexRootHash + ", epoch 5, has value true"
	assert.Equal(t, expectedString, holder.String())
}

func TestNewRootHashHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var holder *rootHashHolder
	require.True(t, holder.IsInterfaceNil())

	holder = NewRootHashHolder([]byte("rootHash"), core.OptionalUint32{})
	require.False(t, holder.IsInterfaceNil())
}
