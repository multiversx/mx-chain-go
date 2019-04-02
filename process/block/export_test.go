package block

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func (bp *blockProcessor) GetTransactionFromPool(senderShardID, destShardID uint32, txHash []byte) *transaction.Transaction {
	return bp.getTransactionFromPool(senderShardID, destShardID, txHash)
}

func (bp *blockProcessor) RequestTransactionFromNetwork(body block.Body) int {
	return bp.requestBlockTransactions(body)
}

func (bp *blockProcessor) WaitForTxHashes(waitTime time.Duration) {
	bp.waitForTxHashes(waitTime)
}

func (bp *blockProcessor) ReceivedTransaction(txHash []byte) {
	bp.receivedTransaction(txHash)
}

func (bp *blockProcessor) ComputeHeaderHash(hdr *block.Header) ([]byte, error) {
	return bp.computeHeaderHash(hdr)
}

func (bp *blockProcessor) DisplayLogInfo(header *block.Header, txBlock block.Body, headerHash []byte) {
	bp.displayLogInfo(header, txBlock, headerHash)
}

func SortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	return sortTxByNonce(txShardStore)
}

func (bp *blockProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bp.verifyStateRoot(rootHash)
}

func (bp *blockProcessor) GetAllTxsFromMiniBlock(mb *block.MiniBlock, haveTime func() bool) ([]*transaction.Transaction, [][]byte, error) {
	return bp.getAllTxsFromMiniBlock(mb, haveTime)
}

func (bp *blockProcessor) ReceivedMiniBlock(miniBlockHash []byte) {
	bp.receivedMiniBlock(miniBlockHash)
}

func (bp *blockProcessor) ReceivedMetaBlock(metaBlockHash []byte) {
	bp.receivedMetaBlock(metaBlockHash)
}

func (bp *blockProcessor) AddTxHashToRequestedList(txHash []byte) {
	bp.mutRequestedTxHashes.Lock()
	defer bp.mutRequestedTxHashes.Unlock()

	if bp.requestedTxHashes == nil {
		bp.requestedTxHashes = make(map[string]bool)
	}
	bp.requestedTxHashes[string(txHash)] = true
}

func (bp *blockProcessor) IsTxHashRequested(txHash []byte) bool {
	bp.mutRequestedTxHashes.Lock()
	defer bp.mutRequestedTxHashes.Unlock()

	_, found := bp.requestedTxHashes[string(txHash)]
	return found
}

func (bp *blockProcessor) ProcessMiniBlockComplete(miniBlock *block.MiniBlock, round int32, haveTime func() bool) error {
	return bp.processMiniBlockComplete(miniBlock, round, haveTime)
}

func (bp *blockProcessor) CreateMiniBlocks(noShards uint32, maxTxInBlock int, round int32, haveTime func() bool) (block.Body, error) {
	return bp.createMiniBlocks(noShards, maxTxInBlock, round, haveTime)
}

func (bp *blockProcessor) RemoveMetaBlockFromPool(blockBody block.Body, blockChain data.ChainHandler) error {
	return bp.removeMetaBlockFromPool(blockBody, blockChain)
}
