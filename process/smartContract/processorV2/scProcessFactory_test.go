package processorV2_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func TestNewSCProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := processorV2.NewSCProcessFactory()
	require.Nil(t, err)
	require.NotNil(t, fact)
	require.Implements(t, new(smartContract.SCProcessorCreator), fact)
}

func TestSCProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	fact, _ := processorV2.NewSCProcessFactory()

	scProcessor, err := fact.CreateSCProcessor(scrCommon.ArgsNewSmartContractProcessor{}, nil)
	require.NotNil(t, err)
	require.Nil(t, scProcessor)

	scProcessor, err = fact.CreateSCProcessor(processorV2.CreateMockSmartContractProcessorArguments(), &epochNotifier.EpochNotifierStub{})
	require.Nil(t, err)
	require.NotNil(t, scProcessor)
	require.Implements(t, new(scrCommon.SCRProcessorHandler), scProcessor)
}

func TestSCProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fact, _ := processorV2.NewSCProcessFactory()
	require.False(t, fact.IsInterfaceNil())
}
