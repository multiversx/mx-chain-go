package processProxy_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/stretchr/testify/require"
)

func TestNewSCProcessProcessFactory(t *testing.T) {
	t.Parallel()

	fact := processProxy.NewSCProcessProxyFactory()
	require.NotNil(t, fact)
	require.Implements(t, new(scrCommon.SCProcessorCreator), fact)
}

func TestSCProcessProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	fact := processProxy.NewSCProcessProxyFactory()

	args := processProxy.CreateMockSmartContractProcessorArguments()
	args.EpochNotifier = nil
	scProcessor, err := fact.CreateSCProcessor(args)
	require.NotNil(t, err)
	require.Nil(t, scProcessor)

	args = processProxy.CreateMockSmartContractProcessorArguments()
	scProcessor, err = fact.CreateSCProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, scProcessor)
	require.Implements(t, new(scrCommon.SCRProcessorHandler), scProcessor)
}

func TestSCProcessProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fact := processProxy.NewSCProcessProxyFactory()
	require.False(t, fact.IsInterfaceNil())
}
