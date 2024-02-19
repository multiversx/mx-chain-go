package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTxPreProcessorCreator(t *testing.T) {
	t.Parallel()

	creator := NewTxPreProcessorCreator()
	require.False(t, creator.IsInterfaceNil())
	require.Implements(t, new(TxPreProcessorCreator), creator)
}

func TestTxPreProcessorCreator_CreateTxPreProcessor(t *testing.T) {
	t.Parallel()

	creator := NewTxPreProcessorCreator()
	args := createDefaultTransactionsProcessorArgs()
	txPreProc, err := creator.CreateTxPreProcessor(args)
	require.NotNil(t, txPreProc)
	require.Nil(t, err)
}
