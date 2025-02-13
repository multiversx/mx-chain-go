package block_test

import (
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
)

func TestNewShardBlockProcessorFactory(t *testing.T) {
	t.Parallel()

	sbpf := block.NewShardBlockProcessorFactory()
	require.NotNil(t, sbpf)
	require.Implements(t, new(block.BlockProcessorCreator), sbpf)
}

func TestShardBlockProcessorFactory_CreateBlockProcessor(t *testing.T) {
	t.Parallel()

	sbpf := block.NewShardBlockProcessorFactory()

	funcCreateMetaArgs := func(systemVM vmcommon.VMExecutionHandler) (*block.ExtraArgsMetaBlockProcessor, error) {
		return nil, nil
	}
	sbp, err := sbpf.CreateBlockProcessor(block.ArgBaseProcessor{}, funcCreateMetaArgs)
	require.NotNil(t, err)
	require.Nil(t, sbp)

	metaArgument := createMockMetaArguments(createMockComponentHolders())
	funcCreateMetaArgs = func(systemVM vmcommon.VMExecutionHandler) (*block.ExtraArgsMetaBlockProcessor, error) {
		return &block.ExtraArgsMetaBlockProcessor{
			EpochStartDataCreator:     metaArgument.EpochStartDataCreator,
			EpochValidatorInfoCreator: metaArgument.EpochValidatorInfoCreator,
			EpochRewardsCreator:       metaArgument.EpochRewardsCreator,
		}, nil
	}
	sbp, err = sbpf.CreateBlockProcessor(metaArgument.ArgBaseProcessor, funcCreateMetaArgs)
	require.NotNil(t, sbp)
	require.Nil(t, err)
	require.Implements(t, new(process.DebuggerBlockProcessor), sbp)
}

func TestShardBlockProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sbpf := block.NewShardBlockProcessorFactory()
	require.False(t, sbpf.IsInterfaceNil())
}
