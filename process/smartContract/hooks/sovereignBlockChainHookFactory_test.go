package hooks

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
)

func TestNewSovereignBlockChainHookFactory(t *testing.T) {
	t.Parallel()

	factory, err := NewSovereignBlockChainHookFactory(nil)

	require.Nil(t, factory)
	require.Equal(t, errors.ErrNilBlockChainHookFactory, err)

	baseFactory := NewBlockChainHookFactory()
	factory, err = NewSovereignBlockChainHookFactory(baseFactory)

	require.Nil(t, err)
	require.NotNil(t, factory)
}

func TestSovereignBlockChainHookFactory_CreateBlockChainHook(t *testing.T) {
	t.Parallel()

	baseFactory := NewBlockChainHookFactory()
	factory, _ := NewSovereignBlockChainHookFactory(baseFactory)

	bhh, err := factory.CreateBlockChainHookHandler(ArgBlockChainHook{})

	require.Nil(t, bhh)
	require.NotNil(t, err)

	bhh, err = factory.CreateBlockChainHookHandler(getDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, bhh)
}

func TestSovereignBlockChainHookFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	baseFactory := NewBlockChainHookFactory()
	factory, _ := NewSovereignBlockChainHookFactory(baseFactory)

	require.False(t, factory.IsInterfaceNil())

	factory = (*sovereignBlockChainHookFactory)(nil)
	require.True(t, factory.IsInterfaceNil())
}
