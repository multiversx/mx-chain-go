package block_test

import (
	"github.com/multiversx/mx-chain-go/testscommon"
	"testing"

	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignBlockProcessorFactory(t *testing.T) {
	t.Parallel()

	sbpf, err := block.NewSovereignBlockProcessorFactory(nil)

	require.Nil(t, sbpf)
	require.NotNil(t, err)

	shardFactory, _ := block.NewShardBlockProcessorFactory()
	sbpf, err = block.NewSovereignBlockProcessorFactory(shardFactory)

	require.NotNil(t, sbpf)
	require.Nil(t, err)
}

func TestSovereignBlockProcessorFactory_CreateBlockProcessor(t *testing.T) {
	t.Parallel()

	shardFactory, _ := block.NewShardBlockProcessorFactory()
	sbpf, _ := block.NewSovereignBlockProcessorFactory(shardFactory)

	sbp, err := sbpf.CreateBlockProcessor(block.ArgBaseProcessor{})
	require.NotNil(t, err)
	require.Nil(t, sbp)

	metaArgument := createMockMetaArguments(createMockComponentHolders())
	metaArgument.ArgBaseProcessor.BlockTracker = &testscommon.ExtendedShardHeaderTrackerStub{}
	metaArgument.ArgBaseProcessor.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}
	sbp, err = sbpf.CreateBlockProcessor(metaArgument.ArgBaseProcessor)

	require.NotNil(t, sbp)
	require.Nil(t, err)
}

func TestSovereignBlockProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	shardFactory, _ := block.NewShardBlockProcessorFactory()
	sbpf, _ := block.NewSovereignBlockProcessorFactory(shardFactory)

	require.False(t, sbpf.IsInterfaceNil())
}
