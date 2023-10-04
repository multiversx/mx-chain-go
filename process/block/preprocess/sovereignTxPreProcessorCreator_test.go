package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignTxPreProcessorCreator(t *testing.T) {
	t.Parallel()

	creator := NewSovereignTxPreProcessorCreator()
	require.False(t, creator.IsInterfaceNil())
	require.Implements(t, new(TxPreProcessorCreator), creator)
}

func TestSovereignTxPreProcessorCreator_CreateTxProcessor(t *testing.T) {
	t.Parallel()

	creator := NewSovereignTxPreProcessorCreator()
	args := createDefaultTransactionsProcessorArgs()
	txProc, err := creator.CreateTxProcessor(args)
	require.NotNil(t, txProc)
	require.Nil(t, err)

	args.Hasher = nil
	txProc, err = creator.CreateTxProcessor(args)
	require.Nil(t, txProc)
	require.NotNil(t, err)
}
