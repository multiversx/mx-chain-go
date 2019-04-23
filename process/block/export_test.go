package block

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func (bp *baseProcessor) ComputeHeaderHash(hdr *block.Header) ([]byte, error) {
	return bp.computeHeaderHash(hdr)
}

func (bp *baseProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bp.verifyStateRoot(rootHash)
}

func (bp *baseProcessor) CheckBlockValidity(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	return bp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
}

func (sp *shardProcessor) GetTransactionFromPool(senderShardID, destShardID uint32, txHash []byte) *transaction.Transaction {
	return sp.getTransactionFromPool(senderShardID, destShardID, txHash)
}

func (sp *shardProcessor) RequestTransactionFromNetwork(body block.Body) int {
	return sp.requestBlockTransactions(body)
}

func (sp *shardProcessor) WaitForTxHashes(waitTime time.Duration) {
	sp.waitForTxHashes(waitTime)
}

func (sp *shardProcessor) ReceivedTransaction(txHash []byte) {
	sp.receivedTransaction(txHash)
}

func (sp *shardProcessor) DisplayShardBlock(header *block.Header, txBlock block.Body) {
	sp.displayShardBlock(header, txBlock)
}

func SortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	return sortTxByNonce(txShardStore)
}

func (sp *shardProcessor) GetAllTxsFromMiniBlock(mb *block.MiniBlock, haveTime func() bool) ([]*transaction.Transaction, [][]byte, error) {
	return sp.getAllTxsFromMiniBlock(mb, haveTime)
}

func (sp *shardProcessor) ReceivedMiniBlock(miniBlockHash []byte) {
	sp.receivedMiniBlock(miniBlockHash)
}

func (sp *shardProcessor) ReceivedMetaBlock(metaBlockHash []byte) {
	sp.receivedMetaBlock(metaBlockHash)
}

func (sp *shardProcessor) AddTxHashToRequestedList(txHash []byte) {
	sp.mutRequestedTxHashes.Lock()
	defer sp.mutRequestedTxHashes.Unlock()

	if sp.requestedTxHashes == nil {
		sp.requestedTxHashes = make(map[string]bool)
	}
	sp.requestedTxHashes[string(txHash)] = true
}

func (sp *shardProcessor) IsTxHashRequested(txHash []byte) bool {
	sp.mutRequestedTxHashes.Lock()
	defer sp.mutRequestedTxHashes.Unlock()

	_, found := sp.requestedTxHashes[string(txHash)]
	return found
}

func (sp *shardProcessor) ProcessMiniBlockComplete(miniBlock *block.MiniBlock, round int32, haveTime func() bool) error {
	return sp.processMiniBlockComplete(miniBlock, round, haveTime)
}

func (sp *shardProcessor) CreateMiniBlocks(noShards uint32, maxTxInBlock int, round int32, haveTime func() bool) (block.Body, error) {
	return sp.createMiniBlocks(noShards, maxTxInBlock, round, haveTime)
}

func (sp *shardProcessor) RemoveMetaBlockFromPool(body block.Body) error {
	return sp.removeMetaBlockFromPool(body)
}

func (sp *shardProcessor) RemoveTxBlockFromPools(blockBody block.Body) error {
	return sp.removeTxBlockFromPools(blockBody)
}
