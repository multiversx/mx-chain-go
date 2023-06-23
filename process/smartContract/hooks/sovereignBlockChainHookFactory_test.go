package hooks

import (
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSovereignBlockChainHookFactory(t *testing.T) {
	t.Parallel()

	factory, err := NewSovereignBlockChainHookFactory(nil)
	assert.Nil(t, factory)
	assert.Equal(t, errors.ErrNilBlockChainHookFactory, err)

	baseFactory, _ := NewBlockChainHookFactory()
	factory, err = NewSovereignBlockChainHookFactory(baseFactory)

	assert.Nil(t, err)
	assert.NotNil(t, factory)
}

func TestSovereignBlockChainHookFactory_CreateBlockChainHook(t *testing.T) {
	t.Parallel()

	baseFactory, _ := NewBlockChainHookFactory()
	factory, _ := NewSovereignBlockChainHookFactory(baseFactory)

	_, err := factory.CreateBlockChainHook(getDefaultArgs())
	assert.Nil(t, err)
}

func TestSovereignBlockChainHookFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	baseFactory, _ := NewBlockChainHookFactory()
	factory, _ := NewSovereignBlockChainHookFactory(baseFactory)

	assert.False(t, factory.IsInterfaceNil())
}
