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

func TestSovereignTxPreProcessorCreator_CreateTxPreProcessor(t *testing.T) {
	t.Parallel()

	creator := NewSovereignTxPreProcessorCreator()
	args := createDefaultTransactionsProcessorArgs()
	txPreProc, err := creator.CreateTxPreProcessor(args)
	require.NotNil(t, txPreProc)
	require.Nil(t, err)

	args.Hasher = nil
	txPreProc, err = creator.CreateTxPreProcessor(args)
	require.Nil(t, txPreProc)
	require.NotNil(t, err)
}
