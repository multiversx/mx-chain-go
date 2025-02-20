package sovereign

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestSovereignPreProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createMockPreProcessorsContainerFactoryArguments()
	sovFactory, err := NewSovereignPreProcessorsContainerFactory(args)
	require.Nil(t, err)
	require.False(t, sovFactory.IsInterfaceNil())

	container, err := sovFactory.Create()
	require.Nil(t, err)
	require.Equal(t, 3, container.Len())

	keys := container.Keys()
	require.Contains(t, keys, block.TxBlock)
	require.Contains(t, keys, block.SmartContractResultBlock)
	require.Contains(t, keys, block.PeerBlock)
}
