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
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedTrieNodeDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.IntMarsh = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	itn, err := NewInterceptedTrieNodeDataFactory(arg)
	assert.Nil(t, itn)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTrieNodeDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	coreComponents.Hash = nil
	arg := createMockArgument(coreComponents, cryptoComponents)

	itn, err := NewInterceptedTrieNodeDataFactory(arg)
	assert.Nil(t, itn)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTrieNodeDataFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents, cryptoComponents := createMockComponentHolders()
	arg := createMockArgument(coreComponents, cryptoComponents)
	itn, err := NewInterceptedTrieNodeDataFactory(arg)
	assert.NotNil(t, itn)
	assert.Nil(t, err)
	assert.False(t, itn.IsInterfaceNil())
}
