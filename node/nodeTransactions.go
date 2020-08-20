package node

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data"
	rewardTxData "github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type transactionType string

const (
	normalTx   transactionType = "normal"
	unsignedTx transactionType = "unsigned"
	rewardTx   transactionType = "reward"
	invalidTx  transactionType = "invalid"
)

// GetTransaction gets the transaction based on the given hash. It will search in the cache and the storage and
// will return the transaction in a format which can be respected by all types of transactions (normal, reward or unsigned)
func (n *Node) GetTransaction(txHash string) (*transaction.ApiTransactionResult, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	if n.historyRepository.IsEnabled() {
		return n.getFullHistoryTransaction(hash)
	}

	return n.getTransaction(hash)
}

func (n *Node) getTransaction(hash []byte) (*transaction.ApiTransactionResult, error) {
	txObj, txType, found := n.getTxObjFromDataPool(hash)
	if found {
		return n.castObjToTransaction(txObj, txType)
	}

	txBytes, txType, found := n.getTxBytesFromStorage(hash)
	if found {
		return n.unmarshalTransaction(txBytes, txType)
	}

	return nil, fmt.Errorf("transaction not found")
}

func (n *Node) getFullHistoryTransaction(hash []byte) (*transaction.ApiTransactionResult, error) {
	// get transaction from pool
	txObj, txType, found := n.getTxObjFromDataPool(hash)
	if found {
		// transaction is in pool return information directly because in history storer
		// there are no additional information about the transaction
		return n.castObjToTransaction(txObj, txType)
	}

	historyTx, err := n.historyRepository.GetTransaction(hash)
	if err != nil {
		// transaction is not in history storer
		return nil, err
	}

	txBytes, txType, found := n.getTxBytesFromStorageByEpoch(hash, historyTx.Epoch)
	if !found {
		// this should never happen because transaction was found in history storer should be also found in transaction
		// storer
		log.Warn("node transaction: cannot find transaction in storage")
		return putHistoryFieldsInTransaction(nil, historyTx), nil
	}

	tx, err := n.unmarshalTransaction(txBytes, txType)
	if err != nil {
		// this should never happen
		log.Warn("node transaction: cannot unmarshal transaction", "error", err.Error())
		return putHistoryFieldsInTransaction(nil, historyTx), nil
	}

	// merge data about transaction from history storer and transaction storer
	tx = putHistoryFieldsInTransaction(tx, historyTx)
	return tx, nil
}

func putHistoryFieldsInTransaction(tx *transaction.ApiTransactionResult, historyFields *fullHistory.HistoryTransactionWithEpoch) *transaction.ApiTransactionResult {
	if tx == nil {
		tx = &transaction.ApiTransactionResult{}
	}

	tx.Epoch = historyFields.Epoch
	tx.MiniBlockHash = hex.EncodeToString(historyFields.MbHash)
	tx.BlockHash = hex.EncodeToString(historyFields.HeaderHash)
	tx.DestinationShard = historyFields.RcvShardID
	tx.SourceShard = historyFields.SndShardID
	tx.Round = historyFields.Round
	tx.BlockNonce = historyFields.HeaderNonce
	return tx
}

func (n *Node) getTxObjFromDataPool(hash []byte) (interface{}, transactionType, bool) {
	txsPool := n.dataPool.Transactions()
	txObj, found := txsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, normalTx, true
	}

	rewardTxsPool := n.dataPool.RewardTransactions()
	txObj, found = rewardTxsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, rewardTx, true
	}

	unsignedTxsPool := n.dataPool.UnsignedTransactions()
	txObj, found = unsignedTxsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, unsignedTx, true
	}

	return nil, invalidTx, false
}

func (n *Node) isTxInStorage(hash []byte) bool {
	txsStorer := n.store.GetStorer(dataRetriever.TransactionUnit)
	err := txsStorer.Has(hash)
	if err == nil {
		return true
	}

	rewardTxsStorer := n.store.GetStorer(dataRetriever.RewardTransactionUnit)
	err = rewardTxsStorer.Has(hash)
	if err == nil {
		return true
	}

	unsignedTxsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	err = unsignedTxsStorer.Has(hash)
	return err == nil
}

func (n *Node) getTxBytesFromStorage(hash []byte) ([]byte, transactionType, bool) {
	txsStorer := n.store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, normalTx, true
	}

	rewardTxsStorer := n.store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, rewardTx, true
	}

	unsignedTxsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, unsignedTx, true
	}

	return nil, invalidTx, false
}

func (n *Node) getTxBytesFromStorageByEpoch(hash []byte, epoch uint32) ([]byte, transactionType, bool) {
	txsStorer := n.store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, normalTx, true
	}

	rewardTxsStorer := n.store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, rewardTx, true
	}

	unsignedTxsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, unsignedTx, true
	}

	return nil, invalidTx, false
}

func (n *Node) castObjToTransaction(txObj interface{}, txType transactionType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case normalTx:
		if tx, ok := txObj.(*transaction.Transaction); ok {
			status := n.computeTransactionStatus(tx, true)
			return n.prepareNormalTx(tx, status)
		}
	case rewardTx:
		if tx, ok := txObj.(*rewardTxData.RewardTx); ok {
			status := n.computeTransactionStatus(tx, true)
			return n.prepareRewardTx(tx, status)
		}
	case unsignedTx:
		if tx, ok := txObj.(*smartContractResult.SmartContractResult); ok {
			status := n.computeTransactionStatus(tx, true)
			return n.prepareUnsignedTx(tx, status)
		}
	}

	return &transaction.ApiTransactionResult{Type: string(invalidTx)}, nil // this shouldn't happen
}

func (n *Node) unmarshalTransaction(txBytes []byte, txType transactionType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case normalTx:
		var tx transaction.Transaction
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		status := n.computeTransactionStatus(&tx, false)
		return n.prepareNormalTx(&tx, status)
	case rewardTx:
		var tx rewardTxData.RewardTx
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		status := n.computeTransactionStatus(&tx, false)
		return n.prepareRewardTx(&tx, status)

	case unsignedTx:
		var tx smartContractResult.SmartContractResult
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		status := n.computeTransactionStatus(&tx, false)
		return n.prepareUnsignedTx(&tx, status)
	default:
		return &transaction.ApiTransactionResult{Type: string(invalidTx)}, nil // this shouldn't happen
	}
}

func (n *Node) prepareNormalTx(tx *transaction.Transaction, status core.TransactionStatus) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Type:      string(normalTx),
		Nonce:     tx.Nonce,
		Value:     tx.Value.String(),
		Receiver:  n.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:    n.addressPubkeyConverter.Encode(tx.SndAddr),
		GasPrice:  tx.GasPrice,
		GasLimit:  tx.GasLimit,
		Data:      tx.Data,
		Signature: hex.EncodeToString(tx.Signature),
		Status:    status,
	}, nil
}

func (n *Node) prepareRewardTx(tx *rewardTxData.RewardTx, status core.TransactionStatus) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Type:     string(rewardTx),
		Round:    tx.GetRound(),
		Epoch:    tx.GetEpoch(),
		Value:    tx.GetValue().String(),
		Sender:   fmt.Sprintf("%d", core.MetachainShardId),
		Receiver: n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		Status:   status,
	}, nil
}

func (n *Node) prepareUnsignedTx(
	tx *smartContractResult.SmartContractResult,
	status core.TransactionStatus,
) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Type:      string(unsignedTx),
		Nonce:     tx.GetNonce(),
		Value:     tx.GetValue().String(),
		Receiver:  n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		Sender:    n.addressPubkeyConverter.Encode(tx.GetSndAddr()),
		GasPrice:  tx.GetGasPrice(),
		GasLimit:  tx.GetGasLimit(),
		Data:      tx.GetData(),
		Code:      string(tx.GetCode()),
		Signature: "",
		Status:    status,
	}, nil
}

func (n *Node) computeTransactionStatus(tx data.TransactionHandler, isInPool bool) core.TransactionStatus {
	selfShardID := n.shardCoordinator.SelfId()
	receiverShardID := n.shardCoordinator.ComputeId(tx.GetRcvAddr())

	var senderShardID uint32
	sndAddr := tx.GetSndAddr()
	if sndAddr != nil {
		senderShardID = n.shardCoordinator.ComputeId(tx.GetSndAddr())
	} else {
		// reward transaction (sender address is nil)
		senderShardID = core.MetachainShardId
	}

	isDestinationMe := selfShardID == receiverShardID
	if isInPool {

		isCrossShard := senderShardID != receiverShardID
		if isDestinationMe && isCrossShard {
			return core.TxStatusPartiallyExecuted
		}

		return core.TxStatusReceived
	}

	// transaction is in storage
	if isDestinationMe {
		return core.TxStatusExecuted
	}

	// is in storage on source shard
	return core.TxStatusPartiallyExecuted
}
