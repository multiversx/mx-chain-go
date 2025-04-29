package trieNodeData

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/stretchr/testify/assert"
)

func TestNewIntermediaryNodeData(t *testing.T) {
	t.Parallel()

	var ind *intermediaryNodeData
	assert.True(t, check.IfNil(ind))

	ind, err := NewIntermediaryNodeData(nil, nil)
	assert.Equal(t, ErrNilKeyBuilder, err)
	assert.True(t, check.IfNil(ind))

	ind, err = NewIntermediaryNodeData(keyBuilder.NewKeyBuilder(), []byte("data"))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(ind))
}

func TestIntermediaryNodeData(t *testing.T) {
	t.Parallel()

	ind, _ := NewIntermediaryNodeData(keyBuilder.NewKeyBuilder(), []byte("data"))
	assert.False(t, ind.IsLeaf())
}
