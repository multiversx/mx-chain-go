package preprocess_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignContractResultPreProcessorFactory(t *testing.T) {
	sovFact, err := preprocess.NewSovereignSmartContractResultPreProcessorFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, sovFact)

	fact, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	sovFact, err = preprocess.NewSovereignSmartContractResultPreProcessorFactory(fact)
	require.Nil(t, err)
	require.NotNil(t, fact)
}

func TestSovereignContractResultPreProcessorFactory_CreateSmartContractResultPreProcessor(t *testing.T) {
	f, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	fact, err := preprocess.NewSovereignSmartContractResultPreProcessorFactory(f)

	args := preprocess.SmartContractResultPreProcessorCreatorArgs{}
	preProcessor, err := fact.CreateSmartContractResultPreProcessor(args)
	require.NotNil(t, err)
	require.Nil(t, preProcessor)

	args = getDefaultSmartContractResultPreProcessorCreatorArgs()
	preProcessor, err = fact.CreateSmartContractResultPreProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, preProcessor)
}

func TestSovereignContractResultPreProcessorFactory_IsInterfaceNil(t *testing.T) {
	f, _ := preprocess.NewSmartContractResultPreProcessorFactory()
	fact, _ := preprocess.NewSovereignSmartContractResultPreProcessorFactory(f)
	require.False(t, fact.IsInterfaceNil())
}
