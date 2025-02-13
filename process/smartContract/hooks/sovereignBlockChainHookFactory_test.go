package hooks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignBlockChainHookFactory(t *testing.T) {
	t.Parallel()

	factory := NewSovereignBlockChainHookFactory()
	require.False(t, factory.IsInterfaceNil())
}

func TestSovereignBlockChainHookFactory_CreateBlockChainHook(t *testing.T) {
	t.Parallel()

	factory := NewSovereignBlockChainHookFactory()
	bhh, err := factory.CreateBlockChainHookHandler(ArgBlockChainHook{})
	require.Nil(t, bhh)
	require.NotNil(t, err)

	bhh, err = factory.CreateBlockChainHookHandler(getDefaultArgs())
	require.Nil(t, err)
	require.Equal(t, "*hooks.sovereignBlockChainHook", fmt.Sprintf("%T", bhh))
}
