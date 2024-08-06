package trieNodeData

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/stretchr/testify/assert"
)

func TestNewLeafNodeData(t *testing.T) {
	t.Parallel()

	var lnd *leafNodeData
	assert.True(t, check.IfNil(lnd))

	lnd, err := NewLeafNodeData(nil, nil)
	assert.Equal(t, ErrNilKeyBuilder, err)
	assert.True(t, check.IfNil(lnd))

	lnd, err = NewLeafNodeData(keyBuilder.NewKeyBuilder(), []byte("data"))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(lnd))
}

func TestLeafNodeData(t *testing.T) {
	t.Parallel()

	lnd, _ := NewLeafNodeData(keyBuilder.NewKeyBuilder(), []byte("data"))
	assert.True(t, lnd.IsLeaf())
}
