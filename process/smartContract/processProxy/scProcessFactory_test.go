package processProxy_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func TestNewSCProcessProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := processProxy.NewSCProcessProxyFactory()
	require.Nil(t, err)
	require.NotNil(t, fact)
	require.Implements(t, new(smartContract.SCProcessorCreator), fact)
}

func TestSCProcessProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	fact, _ := processProxy.NewSCProcessProxyFactory()

	scProcessor, err := fact.CreateSCProcessor(scrCommon.ArgsNewSmartContractProcessor{}, nil)
	require.NotNil(t, err)
	require.Nil(t, scProcessor)

	scProcessor, err = fact.CreateSCProcessor(processProxy.CreateMockSmartContractProcessorArguments(), &epochNotifier.EpochNotifierStub{})
	require.Nil(t, err)
	require.NotNil(t, scProcessor)
	require.Implements(t, new(scrCommon.SCRProcessorHandler), scProcessor)
}

func TestSCProcessProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fact, _ := processProxy.NewSCProcessProxyFactory()
	require.False(t, fact.IsInterfaceNil())
}
