package preprocess_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignContractResultPreProcessorFactory(t *testing.T) {
	t.Parallel()

	sovFact, err := preprocess.NewSovereignSmartContractResultPreProcessorFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, sovFact)

	fact, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	sovFact, err = preprocess.NewSovereignSmartContractResultPreProcessorFactory(fact)
	require.Nil(t, err)
	require.NotNil(t, sovFact)
	require.Implements(t, new(preprocess.SmartContractResultPreProcessorCreator), sovFact)
}

func TestSovereignContractResultPreProcessorFactory_CreateSmartContractResultPreProcessor(t *testing.T) {
	t.Parallel()

	f, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	fact, _ := preprocess.NewSovereignSmartContractResultPreProcessorFactory(f)

	args := preprocess.SmartContractResultPreProcessorCreatorArgs{}
	preProcessor, err := fact.CreateSmartContractResultPreProcessor(args)
	require.NotNil(t, err)
	require.Nil(t, preProcessor)

	args = getDefaultSmartContractResultPreProcessorCreatorArgs()
	preProcessor, err = fact.CreateSmartContractResultPreProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, preProcessor)
	require.Implements(t, new(process.PreProcessor), preProcessor)
}

func TestSovereignContractResultPreProcessorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	f, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	fact, _ := preprocess.NewSovereignSmartContractResultPreProcessorFactory(f)
	require.False(t, fact.IsInterfaceNil())
}
