package chainSimulator

import (
	"testing"

	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/testscommon/chainSimulator"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignProcessorFactory(t *testing.T) {
	t.Parallel()

	fact := NewProcessorFactory()

	require.False(t, fact.IsInterfaceNil())
	require.IsType(t, new(processorFactory), fact)
}

func TestNewSovereignProcessorFactory_CreateChainHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil node handler should error", func(t *testing.T) {
		fact := NewProcessorFactory()

		chainHandler, err := fact.CreateChainHandler(nil)
		require.Nil(t, chainHandler)
		require.ErrorIs(t, err, process.ErrNilNodeHandler)
	})
	t.Run("should work", func(t *testing.T) {
		fact := NewProcessorFactory()

		chainHandler, err := fact.CreateChainHandler(&chainSimulator.NodeHandlerMock{})
		require.Nil(t, err)
		require.NotNil(t, chainHandler)
	})
}
