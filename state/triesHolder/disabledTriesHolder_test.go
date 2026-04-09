package triesHolder

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledDataTriesHolder(t *testing.T) {
	t.Parallel()

	dth := NewDisabledDataTriesHolder()
	assert.False(t, check.IfNil(dth))
}

func TestDisabledDataTriesHolder_PutGetGetAll(t *testing.T) {
	t.Parallel()

	dth := NewDisabledDataTriesHolder()
	dth.Put([]byte("key"), nil)

	trie := dth.Get([]byte("key"))
	assert.Nil(t, trie)
	assert.Empty(t, dth.GetAll())
}

func TestDisabledDataTriesHolder_MarkAsDirtyAndReset(t *testing.T) {
	t.Parallel()

	dth := NewDisabledDataTriesHolder()
	dth.MarkAsDirty([]byte("key"))
	dth.Reset()

	assert.Empty(t, dth.GetAll())
}

func TestDisabledDataTriesHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var dth *disabledDataTriesHolder
	assert.True(t, dth.IsInterfaceNil())

	dth = NewDisabledDataTriesHolder()
	assert.False(t, dth.IsInterfaceNil())
}

