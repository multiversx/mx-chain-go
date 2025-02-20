package processorV2_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/stretchr/testify/require"
)

func TestNewSCProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := processorV2.NewSCProcessFactory()
	require.Nil(t, err)
	require.NotNil(t, fact)
	require.Implements(t, new(scrCommon.SCProcessorCreator), fact)
}

func TestSCProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	t.Run("Nil EpochNotifier should not fail because it is not used", func(t *testing.T) {
		fact, _ := processorV2.NewSCProcessFactory()

		args := processorV2.CreateMockSmartContractProcessorArguments()
		args.EpochNotifier = nil
		scProcessor, err := fact.CreateSCProcessor(args)
		require.Nil(t, err)
		require.NotNil(t, scProcessor)
	})

	t.Run("CreateSCProcessor should work", func(t *testing.T) {
		fact, _ := processorV2.NewSCProcessFactory()

		args := processorV2.CreateMockSmartContractProcessorArguments()
		scProcessor, err := fact.CreateSCProcessor(args)
		require.Nil(t, err)
		require.NotNil(t, scProcessor)
		require.Implements(t, new(scrCommon.SCRProcessorHandler), scProcessor)
	})
}

func TestSCProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fact, _ := processorV2.NewSCProcessFactory()
	require.False(t, fact.IsInterfaceNil())
}
