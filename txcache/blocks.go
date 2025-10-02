package txcache

import "github.com/multiversx/mx-chain-core-go/data/block"

// getTransactionsInBlock returns the transactions from a block body
func getTransactionsInBlock(blockBody *block.Body, txCache txCacheForSelectionTracker) ([]*WrappedTransaction, error) {
	miniBlocks := blockBody.GetMiniBlocks()
	numberOfTxs := computeNumberOfTxsInMiniBlocks(miniBlocks)
	txs := make([]*WrappedTransaction, 0, numberOfTxs)

	for _, miniBlock := range miniBlocks {
		txHashes := miniBlock.GetTxHashes()

		for _, txHash := range txHashes {
			tx, ok := txCache.GetByTxHash(txHash)
			if !ok {
				return nil, errNotFoundTx
			}

			txs = append(txs, tx)
		}
	}

	return txs, nil
}

// computeNumberOfTxsInMiniBlocks returns the number of transactions in mini blocks
func computeNumberOfTxsInMiniBlocks(miniBlocks []*block.MiniBlock) int {
	numberOfTxs := 0
	for _, miniBlock := range miniBlocks {
		numberOfTxs += len(miniBlock.GetTxHashes())
	}

	return numberOfTxs
}
