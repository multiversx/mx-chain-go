package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedTrieNodeDataFactory_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	itn, err := NewInterceptedTrieNodeDataFactory(nil)

	assert.Nil(t, itn)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewInterceptedTrieNodeDataFactory_NilTrieShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Trie = nil

	itn, err := NewInterceptedTrieNodeDataFactory(arg)
	assert.Nil(t, itn)
	assert.Equal(t, process.ErrNilTrie, err)
}

func TestNewInterceptedTrieNodeDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	itn, err := NewInterceptedTrieNodeDataFactory(arg)
	assert.Nil(t, itn)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTrieNodeDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	itn, err := NewInterceptedTrieNodeDataFactory(arg)
	assert.Nil(t, itn)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTrieNodeDataFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	itn, err := NewInterceptedTrieNodeDataFactory(createMockArgument())
	assert.NotNil(t, itn)
	assert.Nil(t, err)
}
