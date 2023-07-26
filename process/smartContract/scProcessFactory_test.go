package smartContract_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/stretchr/testify/require"
)

func TestNewSCProcessFactory(t *testing.T) {
	t.Parallel()

	fact, err := smartContract.NewSCProcessFactory()
	require.Nil(t, err)
	require.NotNil(t, fact)
}

func TestSCProcessFactory_CreateSCProcessor(t *testing.T) {
	t.Parallel()

	fact, _ := smartContract.NewSCProcessFactory()

	scProcessor, err := fact.CreateSCProcessor(smartContract.ArgsNewSmartContractProcessor{})
	require.NotNil(t, err)
	require.Nil(t, scProcessor)

	scProcessor, err = fact.CreateSCProcessor(smartContract.CreateMockSmartContractProcessorArguments())
	require.Nil(t, err)
	require.NotNil(t, scProcessor)
}

func TestSCProcessFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	fact, _ := smartContract.NewSCProcessFactory()
	require.False(t, fact.IsInterfaceNil())
}
