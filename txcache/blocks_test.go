package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func Test_computeNumberOfTxsInMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("should return the right number of txs", func(t *testing.T) {
		blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash1"),
					[]byte("txHash2"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash4"),
					[]byte("txHash5"),
					[]byte("txHash6"),
				},
			},
		}}

		actualResult := computeNumberOfTxsInMiniBlocks(blockBody.MiniBlocks)
		require.Equal(t, 6, actualResult)
	})
}

func Test_getTransactionsInBlock(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash1"),
					[]byte("txHash2"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		}}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash = newTxByHashMap(1)

		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2))
		txCache.txByHash.addTx(createTx([]byte("txHash3"), "alice", 3))

		txs, err := getTransactionsInBlock(&blockBody, txCache, 0)
		require.Nil(t, err)
		require.Equal(t, 3, len(txs))
	})

	t.Run("should fail", func(t *testing.T) {
		blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash1"),
					[]byte("txHash2"),
				},
			},
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		}}

		txCache := newCacheToTest(maxNumBytesPerSenderUpperBoundTest, 3)
		txCache.txByHash = newTxByHashMap(1)

		txCache.txByHash.addTx(createTx([]byte("txHash1"), "alice", 1))
		txCache.txByHash.addTx(createTx([]byte("txHash2"), "alice", 2))

		txs, err := getTransactionsInBlock(&blockBody, txCache, 0)
		require.Nil(t, txs)
		require.Equal(t, errNotFoundTx, err)
	})
}
