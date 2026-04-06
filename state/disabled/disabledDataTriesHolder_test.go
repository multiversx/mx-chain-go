package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledDataTriesHolder(t *testing.T) {
	t.Parallel()

	ddth := NewDisabledDataTriesHolder()

	assert.False(t, check.IfNil(ddth))
	assert.Nil(t, ddth.Get([]byte("key")))
	assert.Empty(t, ddth.GetAll())

	ddth.Put([]byte("key"), &trieMock.TrieStub{})
	ddth.MarkAsDirty([]byte("key"))
	ddth.Reset()

	assert.Nil(t, ddth.Get([]byte("key")))
	assert.Empty(t, ddth.GetAll())
}
