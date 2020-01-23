package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilFinalityAttester_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var nfa *nilFinalityAttester
	assert.True(t, check.IfNil(nfa))

	nfa = &nilFinalityAttester{}
	assert.False(t, check.IfNil(nfa))
}

func TestNilFinalityAttester_GetFinalHeader(t *testing.T) {
	t.Parallel()

	nfa := &nilFinalityAttester{}
	shardId := uint32(8)
	hdr, hash, err := nfa.GetFinalHeader(shardId)

	require.Nil(t, err)
	require.False(t, check.IfNil(hdr))
	assert.Equal(t, shardId, hdr.GetShardID())
	assert.Equal(t, uint64(0), hdr.GetNonce())
	assert.Equal(t, 0, len(hash))
}
