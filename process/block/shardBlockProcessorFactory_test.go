package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/stretchr/testify/require"
)

func TestNewShardBlockProcessorFactory(t *testing.T) {
	t.Parallel()

	sbpf, err := block.NewShardBlockProcessorFactory()
	require.NotNil(t, sbpf)
	require.Nil(t, err)
	require.Implements(t, new(block.BlockProcessorCreator), sbpf)
}

func TestShardBlockProcessorFactory_CreateBlockProcessor(t *testing.T) {
	t.Parallel()

	sbpf, _ := block.NewShardBlockProcessorFactory()
	sbp, err := sbpf.CreateBlockProcessor(block.ArgBaseProcessor{})
	require.NotNil(t, err)
	require.Nil(t, sbp)

	metaArgument := createMockMetaArguments(createMockComponentHolders())
	sbp, err = sbpf.CreateBlockProcessor(metaArgument.ArgBaseProcessor)
	require.NotNil(t, sbp)
	require.Nil(t, err)
	require.Implements(t, new(process.DebuggerBlockProcessor), sbp)
}

func TestShardBlockProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sbpf, _ := block.NewShardBlockProcessorFactory()
	require.False(t, sbpf.IsInterfaceNil())
}
