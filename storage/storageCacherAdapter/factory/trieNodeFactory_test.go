package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieNodeFactory(t *testing.T) {
	t.Parallel()

	tnf := NewTrieNodeFactory()
	assert.False(t, check.IfNil(tnf))
}

func TestTrieNodeFactory_CreateEmpty(t *testing.T) {
	t.Parallel()

	tnf := NewTrieNodeFactory()

	emptyInterceptedNode := tnf.CreateEmpty()
	n, ok := emptyInterceptedNode.(*trie.InterceptedTrieNode)
	assert.True(t, ok)
	assert.NotNil(t, n)
}
