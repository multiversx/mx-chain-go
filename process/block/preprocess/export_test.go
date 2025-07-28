package preprocess

import (
	"math/big"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

func (txs *transactions) ReceivedTransaction(txHash []byte, value interface{}) {
	txs.receivedTransaction(txHash, value)
}

func (txs *transactions) AddTxHashToRequestedList(txHash []byte) {
	txsForCurrentBlock := txs.txsForCurrBlock.(*txsForBlock)
	txsForCurrentBlock.mutTxsForBlock.Lock()
	defer txsForCurrentBlock.mutTxsForBlock.Unlock()

	if txsForCurrentBlock.txHashAndInfo == nil {
		txsForCurrentBlock.txHashAndInfo = make(map[string]*txInfo)
	}
	txsForCurrentBlock.txHashAndInfo[string(txHash)] = &txInfo{txShardInfo: &txShardInfo{}}
}

func (txs *transactions) IsTxHashRequested(txHash []byte) bool {
	txsForCurrentBlock := txs.txsForCurrBlock.(*txsForBlock)
	txsForCurrentBlock.mutTxsForBlock.Lock()
	defer txsForCurrentBlock.mutTxsForBlock.Unlock()

	return txsForCurrentBlock.txHashAndInfo[string(txHash)].tx == nil ||
		txsForCurrentBlock.txHashAndInfo[string(txHash)].tx.IsInterfaceNil()
}

func (txs *transactions) SetMissingTxs(missingTxs int) {
	txsForCurrentBlock := txs.txsForCurrBlock.(*txsForBlock)

	txsForCurrentBlock.mutTxsForBlock.Lock()
	txsForCurrentBlock.missingTxs = missingTxs
	txsForCurrentBlock.mutTxsForBlock.Unlock()
}

func (txs *transactions) SetRcvdTxChan() {
	txs.chRcvAllTxs <- true
}

func (scr *smartContractResults) AddScrHashToRequestedList(txHash []byte) {
	scrForBlock := scr.scrForBlock.(*txsForBlock)
	scrForBlock.mutTxsForBlock.Lock()
	defer scrForBlock.mutTxsForBlock.Unlock()

	if scrForBlock.txHashAndInfo == nil {
		scrForBlock.txHashAndInfo = make(map[string]*txInfo)
	}
	scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{txShardInfo: &txShardInfo{}}
}

func (scr *smartContractResults) IsScrHashRequested(txHash []byte) bool {
	scrForBlock := scr.scrForBlock.(*txsForBlock)
	scrForBlock.mutTxsForBlock.Lock()
	defer scrForBlock.mutTxsForBlock.Unlock()

	return scrForBlock.txHashAndInfo[string(txHash)].tx == nil ||
		scrForBlock.txHashAndInfo[string(txHash)].tx.IsInterfaceNil()
}

func (scr *smartContractResults) SetMissingScr(missingTxs int) {
	missingScrsForBlock, _ := scr.scrForBlock.(*txsForBlock)
	missingScrsForBlock.mutTxsForBlock.Lock()
	missingScrsForBlock.missingTxs = missingTxs
	missingScrsForBlock.mutTxsForBlock.Unlock()
}

func (rtp *rewardTxPreprocessor) AddTxs(txHashes [][]byte, txs []data.TransactionHandler) {
	for i := 0; i < len(txHashes); i++ {
		rtp.rewardTxsForBlock.AddTransaction(txHashes[i], txs[i], core.MetachainShardId, 0)
	}
}

func (bsc *blockSizeComputation) MiniblockSize() uint32 {
	return bsc.miniblockSize
}

func (bsc *blockSizeComputation) TxSize() uint32 {
	return bsc.txSize
}

func (bsc *blockSizeComputation) NumMiniBlocks() uint32 {
	return atomic.LoadUint32(&bsc.numMiniBlocks)
}

func (bsc *blockSizeComputation) NumTxs() uint32 {
	return atomic.LoadUint32(&bsc.numTxs)
}

func (txs *transactions) ProcessTxsToMe(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) error {
	return txs.processTxsToMe(header, body, haveTime)
}

func (txs *transactions) AddTxForCurrentBlock(
	txHash []byte,
	txHandler data.TransactionHandler,
	senderShardID uint32,
	receiverShardID uint32,
) {
	txs.txsForCurrBlock.AddTransaction(txHash, txHandler, senderShardID, receiverShardID)
}

func (txs *transactions) GetTxInfoForCurrentBlock(txHash []byte) (data.TransactionHandler, uint32, uint32) {
	txInfo, ok := txs.txsForCurrBlock.GetTxInfoByHash(txHash)
	if !ok {
		return nil, 0, 0
	}

	return txInfo.tx, txInfo.senderShardID, txInfo.receiverShardID
}

func (bc *balanceComputation) GetBalanceOfAddress(address []byte) *big.Int {
	bc.mutAddressBalance.RLock()
	defer bc.mutAddressBalance.RUnlock()

	currValue, ok := bc.mapAddressBalance[string(address)]
	if !ok {
		return nil
	}

	return big.NewInt(0).Set(currValue)
}

func (gc *gasComputation) GetTxHashesWithGasProvidedSinceLastReset(key []byte) [][]byte {
	return gc.getTxHashesWithGasProvidedSinceLastReset(key)
}

func (gc *gasComputation) GetTxHashesWithGasProvidedAsScheduledSinceLastReset(key []byte) [][]byte {
	return gc.getTxHashesWithGasProvidedAsScheduledSinceLastReset(key)
}

func (gc *gasComputation) GetTxHashesWithGasRefundedSinceLastReset(key []byte) [][]byte {
	return gc.getTxHashesWithGasRefundedSinceLastReset(key)
}

func (gc *gasComputation) GetTxHashesWithGasPenalizedSinceLastReset(key []byte) [][]byte {
	return gc.getTxHashesWithGasPenalizedSinceLastReset(key)
}

func (ste *scheduledTxsExecution) ComputeScheduledIntermediateTxs(
	mapAllIntermediateTxsBeforeScheduledExecution map[block.Type]map[string]data.TransactionHandler,
	mapAllIntermediateTxsAfterScheduledExecution map[block.Type]map[string]data.TransactionHandler,
) {
	ste.mutScheduledTxs.Lock()
	ste.computeScheduledIntermediateTxs(mapAllIntermediateTxsBeforeScheduledExecution, mapAllIntermediateTxsAfterScheduledExecution)
	ste.mutScheduledTxs.Unlock()
}

func (ste *scheduledTxsExecution) GetMapScheduledIntermediateTxs() map[block.Type][]data.TransactionHandler {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	newMap := make(map[block.Type][]data.TransactionHandler)
	for key, value := range ste.mapScheduledIntermediateTxs {
		newMap[key] = value
	}

	return newMap
}
