package smartContract_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignSCProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := smartContract.NewSovereignSCProcessFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, fact)

	f, _ := smartContract.NewSCProcessFactory()
	fact, err = smartContract.NewSovereignSCProcessFactory(f)
	require.Nil(t, err)
	require.NotNil(t, fact)
}

func TestSovereignSCProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	f, _ := smartContract.NewSCProcessFactory()
	fact, _ := smartContract.NewSovereignSCProcessFactory(f)

	scProcessor, err := fact.CreateSCProcessor(smartContract.ArgsNewSmartContractProcessor{})
	require.NotNil(t, err)
	require.Nil(t, scProcessor)

	scProcessor, err = fact.CreateSCProcessor(smartContract.CreateMockSmartContractProcessorArguments())
	require.Nil(t, err)
	require.NotNil(t, scProcessor)
}

func TestSovereignSCProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	f, _ := smartContract.NewSCProcessFactory()
	fact, _ := smartContract.NewSovereignSCProcessFactory(f)
	require.False(t, fact.IsInterfaceNil())
}
