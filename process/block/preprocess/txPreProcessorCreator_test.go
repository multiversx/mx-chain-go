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

func TestTxPreProcessorCreator_CreateTxProcessor(t *testing.T) {
	t.Parallel()

	creator := NewTxPreProcessorCreator()
	args := createDefaultTransactionsProcessorArgs()
	txProc, err := creator.CreateTxProcessor(args)
	require.NotNil(t, txProc)
	require.Nil(t, err)
}
