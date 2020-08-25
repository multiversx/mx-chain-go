package node

import (
	"bytes"
	"encoding/hex"

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

	tx, err := n.optionallyGetTransactionFromPool(hash)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		return tx, nil
	}

	if n.historyRepository.IsEnabled() {
		return n.getFullHistoryTransaction(hash)
	}

	return n.getTransactionFromCurrentEpochStorage(hash)
}

func (n *Node) optionallyGetTransactionFromPool(hash []byte) (*transaction.ApiTransactionResult, error) {
	txObj, txType, found := n.getTxObjFromDataPool(hash)
	if !found {
		return nil, nil
	}

	tx, err := n.castObjToTransaction(txObj, txType)
	if err != nil {
		return nil, err
	}

	tx.Status = transaction.ComputeStatusWhenInPool(tx.SourceShard, tx.DestinationShard, n.shardCoordinator.SelfId())
	return tx, nil
}

func (n *Node) getFullHistoryTransaction(hash []byte) (*transaction.ApiTransactionResult, error) {
	miniblockMetadata, err := n.historyRepository.GetMiniblockMetadataByTxHash(hash)
	if err != nil {
		return nil, ErrTransactionNotFound
	}

	txBytes, txType, found := n.getTxBytesFromStorageByEpoch(hash, miniblockMetadata.Epoch)
	if !found {
		log.Warn("getFullHistoryTransaction(): unexpected condition, cannot find transaction in storage")
		return nil, ErrCannotRetrieveTransaction
	}

	tx, err := n.unmarshalTransaction(txBytes, txType)
	if err != nil {
		log.Warn("getFullHistoryTransaction(): unexpected condition, cannot unmarshal transaction")
		return nil, ErrCannotRetrieveTransaction
	}

	putHistoryFieldsInTransaction(tx, miniblockMetadata)
	return tx, nil
}

func putHistoryFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *fullHistory.MiniblockMetadata) *transaction.ApiTransactionResult {
	tx.Epoch = miniblockMetadata.Epoch
	tx.Round = miniblockMetadata.Round

	tx.MiniBlockHash = hex.EncodeToString(miniblockMetadata.MiniblockHash)
	tx.DestinationShard = miniblockMetadata.DestinationShardID
	tx.SourceShard = miniblockMetadata.SourceShardID

	tx.BlockNonce = miniblockMetadata.HeaderNonce
	tx.BlockHash = hex.EncodeToString(miniblockMetadata.HeaderHash)
	tx.NotarizedAtSourceInMetaNonce = miniblockMetadata.NotarizedAtSourceInMetaNonce
	tx.NotarizedAtSourceInMetaHash = hex.EncodeToString(miniblockMetadata.NotarizedAtSourceInMetaHash)
	tx.NotarizedAtDestinationInMetaNonce = miniblockMetadata.NotarizedAtDestinationInMetaNonce
	tx.NotarizedAtDestinationInMetaHash = hex.EncodeToString(miniblockMetadata.NotarizedAtDestinationInMetaHash)
	tx.Status = transaction.TxStatus(miniblockMetadata.Status)

	return tx
}

func (n *Node) getTransactionFromCurrentEpochStorage(hash []byte) (*transaction.ApiTransactionResult, error) {
	txBytes, txType, found := n.getTxBytesFromStorage(hash)
	if !found {
		return nil, ErrTransactionNotFound
	}

	tx, err := n.unmarshalTransaction(txBytes, txType)
	if err != nil {
		return nil, err
	}

	tx.Status = transaction.ComputeStatusWhenInStorage(tx.SourceShard, tx.DestinationShard, n.shardCoordinator.SelfId())
	return tx, nil
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
			return n.prepareNormalTx(tx)
		}
	case rewardTx:
		if tx, ok := txObj.(*rewardTxData.RewardTx); ok {
			return n.prepareRewardTx(tx)
		}
	case unsignedTx:
		if tx, ok := txObj.(*smartContractResult.SmartContractResult); ok {
			return n.prepareUnsignedTx(tx)
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
		return n.prepareNormalTx(&tx)
	case rewardTx:
		var tx rewardTxData.RewardTx
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return n.prepareRewardTx(&tx)

	case unsignedTx:
		var tx smartContractResult.SmartContractResult
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return n.prepareUnsignedTx(&tx)
	default:
		return &transaction.ApiTransactionResult{Type: string(invalidTx)}, nil // this shouldn't happen
	}
}

func (n *Node) prepareNormalTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	sourceShard := n.shardCoordinator.ComputeId(tx.GetSndAddr())
	destinationShard := n.shardCoordinator.ComputeId(tx.GetRcvAddr())

	return &transaction.ApiTransactionResult{
		Type:             string(normalTx),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         n.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:           n.addressPubkeyConverter.Encode(tx.SndAddr),
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
		SourceShard:      sourceShard,
		DestinationShard: destinationShard,
	}, nil
}

func (n *Node) prepareRewardTx(tx *rewardTxData.RewardTx) (*transaction.ApiTransactionResult, error) {
	destinationShard := n.shardCoordinator.ComputeId(tx.GetRcvAddr())

	return &transaction.ApiTransactionResult{
		Type:             string(rewardTx),
		Round:            tx.GetRound(),
		Epoch:            tx.GetEpoch(),
		Value:            tx.GetValue().String(),
		Sender:           "metachain",
		Receiver:         n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		SourceShard:      core.MetachainShardId,
		DestinationShard: destinationShard,
	}, nil
}

func (n *Node) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	sourceShard := n.shardCoordinator.ComputeId(tx.GetSndAddr())
	destinationShard := n.shardCoordinator.ComputeId(tx.GetRcvAddr())

	return &transaction.ApiTransactionResult{
		Type:             string(unsignedTx),
		Nonce:            tx.GetNonce(),
		Value:            tx.GetValue().String(),
		Receiver:         n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		Sender:           n.addressPubkeyConverter.Encode(tx.GetSndAddr()),
		GasPrice:         tx.GetGasPrice(),
		GasLimit:         tx.GetGasLimit(),
		Data:             tx.GetData(),
		Code:             string(tx.GetCode()),
		SourceShard:      sourceShard,
		DestinationShard: destinationShard,
	}, nil
}

func (n *Node) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRcvAddr(), make([]byte, n.addressPubkeyConverter.Len()))
	return isEmptyAddress
}
