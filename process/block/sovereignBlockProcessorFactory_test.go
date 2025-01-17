package block_test

import (
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewSovereignBlockProcessorFactory(t *testing.T) {
	t.Parallel()

	sbpf, err := block.NewSovereignBlockProcessorFactory(nil)

	require.Nil(t, sbpf)
	require.NotNil(t, err)

	shardFactory := block.NewShardBlockProcessorFactory()
	sbpf, err = block.NewSovereignBlockProcessorFactory(shardFactory)

	require.NotNil(t, sbpf)
	require.Nil(t, err)
	require.Implements(t, new(block.BlockProcessorCreator), sbpf)
}

func TestSovereignBlockProcessorFactory_CreateBlockProcessor(t *testing.T) {
	t.Parallel()

	shardFactory := block.NewShardBlockProcessorFactory()
	sbpf, _ := block.NewSovereignBlockProcessorFactory(shardFactory)

	funcCreateMetaArgs := func(systemVM vmcommon.VMExecutionHandler) (*block.ExtraArgsMetaBlockProcessor, error) {
		return nil, nil
	}
	sbp, err := sbpf.CreateBlockProcessor(block.ArgBaseProcessor{}, funcCreateMetaArgs)
	require.NotNil(t, err)
	require.Nil(t, sbp)

	metaArgument := createMockMetaArguments(createMockComponentHolders())
	metaArgument.ArgBaseProcessor.BlockTracker = &testscommon.ExtendedShardHeaderTrackerStub{}
	metaArgument.ArgBaseProcessor.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}
	metaArgument.ArgBaseProcessor.Config = testscommon.GetGeneralConfig()

	funcCreateMetaArgs = func(systemVM vmcommon.VMExecutionHandler) (*block.ExtraArgsMetaBlockProcessor, error) {
		return &block.ExtraArgsMetaBlockProcessor{
			EpochStartDataCreator:     metaArgument.EpochStartDataCreator,
			EpochValidatorInfoCreator: metaArgument.EpochValidatorInfoCreator,
			EpochRewardsCreator:       metaArgument.EpochRewardsCreator,
			EpochSystemSCProcessor:    metaArgument.EpochSystemSCProcessor,
			EpochEconomics:            &mock.EpochEconomicsStub{},
			SCToProtocol:              &mock.SCToProtocolStub{},
		}, nil
	}

	sbp, err = sbpf.CreateBlockProcessor(metaArgument.ArgBaseProcessor, funcCreateMetaArgs)
	require.Nil(t, err)
	require.NotNil(t, sbp)
	require.Implements(t, new(process.BlockProcessor), sbp)
}

func TestSovereignBlockProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	shardFactory := block.NewShardBlockProcessorFactory()
	sbpf, _ := block.NewSovereignBlockProcessorFactory(shardFactory)
	require.False(t, sbpf.IsInterfaceNil())
}
