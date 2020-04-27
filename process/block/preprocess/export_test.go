package preprocess

import (
	"math/big"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

func (txs *transactions) ReceivedTransaction(txHash []byte, value interface{}) {
	txs.receivedTransaction(txHash, value)
}

func (txs *transactions) AddTxHashToRequestedList(txHash []byte) {
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	defer txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	if txs.txsForCurrBlock.txHashAndInfo == nil {
		txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	}
	txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{txShardInfo: &txShardInfo{}}
}

func (txs *transactions) IsTxHashRequested(txHash []byte) bool {
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	defer txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return txs.txsForCurrBlock.txHashAndInfo[string(txHash)].tx == nil ||
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)].tx.IsInterfaceNil()
}

func (txs *transactions) SetMissingTxs(missingTxs int) {
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.missingTxs = missingTxs
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()
}

func (txs *transactions) SetRcvdTxChan() {
	txs.chRcvAllTxs <- true
}

func (scr *smartContractResults) AddScrHashToRequestedList(txHash []byte) {
	scr.scrForBlock.mutTxsForBlock.Lock()
	defer scr.scrForBlock.mutTxsForBlock.Unlock()

	if scr.scrForBlock.txHashAndInfo == nil {
		scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)
	}
	scr.scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{txShardInfo: &txShardInfo{}}
}

func (scr *smartContractResults) IsScrHashRequested(txHash []byte) bool {
	scr.scrForBlock.mutTxsForBlock.Lock()
	defer scr.scrForBlock.mutTxsForBlock.Unlock()

	return scr.scrForBlock.txHashAndInfo[string(txHash)].tx == nil ||
		scr.scrForBlock.txHashAndInfo[string(txHash)].tx.IsInterfaceNil()
}

func (scr *smartContractResults) SetMissingScr(missingTxs int) {
	scr.scrForBlock.mutTxsForBlock.Lock()
	scr.scrForBlock.missingTxs = missingTxs
	scr.scrForBlock.mutTxsForBlock.Unlock()
}

func (rtp *rewardTxPreprocessor) AddTxs(txHashes [][]byte, txs []data.TransactionHandler) {
	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()

	for i := 0; i < len(txHashes); i++ {
		hash := txHashes[i]
		tx := txs[i]
		rtp.rewardTxsForBlock.txHashAndInfo[string(hash)] = &txInfo{
			tx:          tx,
			txShardInfo: &txShardInfo{receiverShardID: core.MetachainShardId, senderShardID: 0},
		}
	}
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()
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
	body *block.Body,
	haveTime func() bool,
) error {
	return txs.processTxsToMe(body, haveTime)
}

func (txs *transactions) AddTxForCurrentBlock(
	txHash []byte,
	txHandler data.TransactionHandler,
	senderShardID uint32,
	receiverShardID uint32,
) {
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	defer txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	if txs.txsForCurrBlock.txHashAndInfo == nil {
		txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	}

	txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{
		tx: txHandler,
		txShardInfo: &txShardInfo{
			senderShardID:   senderShardID,
			receiverShardID: receiverShardID,
		},
	}
}

func (txs *transactions) GetTxInfoForCurrentBlock(txHash []byte) (data.TransactionHandler, uint32, uint32) {
	txs.txsForCurrBlock.mutTxsForBlock.RLock()
	defer txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

	txInfo, ok := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
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
