package processorV2_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignSCProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := processorV2.NewSovereignSCProcessFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, fact)

	f, _ := processorV2.NewSCProcessFactory()
	fact, err = processorV2.NewSovereignSCProcessFactory(f)
	require.Nil(t, err)
	require.NotNil(t, fact)
	require.Implements(t, new(scrCommon.SCProcessorCreator), fact)
}

func TestSovereignSCProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	f, _ := processorV2.NewSCProcessFactory()
	fact, _ := processorV2.NewSovereignSCProcessFactory(f)

	scProcessor, err := fact.CreateSCProcessor(scrCommon.ArgsNewSmartContractProcessor{}, nil)
	require.NotNil(t, err)
	require.Nil(t, scProcessor)

	scProcessor, err = fact.CreateSCProcessor(processorV2.CreateMockSmartContractProcessorArguments(), nil)
	require.Nil(t, err)
	require.NotNil(t, scProcessor)
	require.Implements(t, new(scrCommon.SCRProcessorHandler), scProcessor)
}

func TestSovereignSCProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	f, _ := processorV2.NewSCProcessFactory()
	fact, _ := processorV2.NewSovereignSCProcessFactory(f)
	require.False(t, fact.IsInterfaceNil())
}
