package dataPool

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestCurrentBlockPool_AddGetCleanTx(t *testing.T) {
	t.Parallel()

	txHash := []byte("hash")
	tx := &transaction.Transaction{}
	currentBlockPool := NewCurrentBlockTransactionsPool()
	require.False(t, currentBlockPool.IsInterfaceNil())

	currentBlockPool.AddTx(txHash, tx)
	currentBlockPool.AddTx(txHash, nil)

	txFromPool, err := currentBlockPool.GetTx([]byte("wrong hash"))
	require.Nil(t, txFromPool)
	require.Equal(t, dataRetriever.ErrTxNotFoundInBlockPool, err)

	txFromPool, err = currentBlockPool.GetTx(txHash)
	require.Nil(t, err)
	require.Equal(t, tx, txFromPool)

	currentBlockPool.Clean()
	txFromPool, err = currentBlockPool.GetTx(txHash)
	require.Nil(t, txFromPool)
	require.Equal(t, dataRetriever.ErrTxNotFoundInBlockPool, err)
}
