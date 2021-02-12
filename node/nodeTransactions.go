package node

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/block"
	rewardTxData "github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// GetTransaction gets the transaction based on the given hash. It will search in the cache and the storage and
// will return the transaction in a format which can be respected by all types of transactions (normal, reward or unsigned)
func (n *Node) GetTransaction(txHash string, withResults bool) (*transaction.ApiTransactionResult, error) {
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
		return n.lookupHistoricalTransaction(hash, withResults)
	}

	return n.getTransactionFromStorage(hash)
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

	tx.Status = transaction.TxStatusPending

	return tx, nil
}

func (n *Node) lookupHistoricalTransaction(hash []byte, withResults bool) (*transaction.ApiTransactionResult, error) {
	miniblockMetadata, err := n.historyRepository.GetMiniblockMetadataByTxHash(hash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrTransactionNotFound.Error(), err)
	}

	txBytes, txType, found := n.getTxBytesFromStorageByEpoch(hash, miniblockMetadata.Epoch)
	if !found {
		log.Warn("lookupHistoricalTransaction(): unexpected condition, cannot find transaction in storage")
		return nil, fmt.Errorf("%s: %w", ErrCannotRetrieveTransaction.Error(), err)
	}

	// After looking up a transaction from storage, it's impossible to say whether it was successful or invalid
	// (since both successful and invalid transactions are kept in the same storage unit),
	// so we have to use our extra information from the "miniblockMetadata" to correct the txType if appropriate
	if block.Type(miniblockMetadata.Type) == block.InvalidBlock {
		txType = transaction.TxTypeInvalid
	}

	tx, err := n.unmarshalTransaction(txBytes, txType)
	if err != nil {
		log.Warn("lookupHistoricalTransaction(): unexpected condition, cannot unmarshal transaction")
		return nil, fmt.Errorf("%s: %w", ErrCannotRetrieveTransaction.Error(), err)
	}

	putMiniblockFieldsInTransaction(tx, miniblockMetadata)
	statusComputer, err := transaction.NewStatusComputer(n.shardCoordinator.SelfId(), n.uint64ByteSliceConverter, n.store)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrNilStatusComputer.Error(), err)
	}

	if ok, _ := statusComputer.SetStatusIfIsRewardReverted(
		tx,
		block.Type(miniblockMetadata.Type),
		miniblockMetadata.HeaderNonce,
		miniblockMetadata.HeaderHash); ok {
		return tx, nil
	}

	tx.Status, _ = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(
		block.Type(miniblockMetadata.Type), tx)

	if withResults {
		n.putResultsInTransaction(hash, tx, miniblockMetadata.Epoch)
	}

	return tx, nil
}

func putMiniblockFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *dblookupext.MiniblockMetadata) *transaction.ApiTransactionResult {
	tx.Epoch = miniblockMetadata.Epoch
	tx.Round = miniblockMetadata.Round

	tx.MiniBlockType = block.Type(miniblockMetadata.Type).String()
	tx.MiniBlockHash = hex.EncodeToString(miniblockMetadata.MiniblockHash)
	tx.DestinationShard = miniblockMetadata.DestinationShardID
	tx.SourceShard = miniblockMetadata.SourceShardID

	tx.BlockNonce = miniblockMetadata.HeaderNonce
	tx.BlockHash = hex.EncodeToString(miniblockMetadata.HeaderHash)
	tx.NotarizedAtSourceInMetaNonce = miniblockMetadata.NotarizedAtSourceInMetaNonce
	tx.NotarizedAtSourceInMetaHash = hex.EncodeToString(miniblockMetadata.NotarizedAtSourceInMetaHash)
	tx.NotarizedAtDestinationInMetaNonce = miniblockMetadata.NotarizedAtDestinationInMetaNonce
	tx.NotarizedAtDestinationInMetaHash = hex.EncodeToString(miniblockMetadata.NotarizedAtDestinationInMetaHash)

	return tx
}

func (n *Node) getTransactionFromStorage(hash []byte) (*transaction.ApiTransactionResult, error) {
	txBytes, txType, found := n.getTxBytesFromStorage(hash)
	if !found {
		return nil, ErrTransactionNotFound
	}

	tx, err := n.unmarshalTransaction(txBytes, txType)
	if err != nil {
		return nil, err
	}

	// TODO: take care of this when integrating the adaptivity
	statusComputer, err := transaction.NewStatusComputer(n.shardCoordinator.SelfId(), n.uint64ByteSliceConverter, n.store)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrNilStatusComputer.Error(), err)
	}
	tx.Status, _ = statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(
		n.shardCoordinator.ComputeId(tx.Tx.GetRcvAddr()), tx)

	return tx, nil
}

func (n *Node) getTxObjFromDataPool(hash []byte) (interface{}, transaction.TxType, bool) {
	txsPool := n.dataPool.Transactions()
	txObj, found := txsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeNormal, true
	}

	rewardTxsPool := n.dataPool.RewardTransactions()
	txObj, found = rewardTxsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeReward, true
	}

	unsignedTxsPool := n.dataPool.UnsignedTransactions()
	txObj, found = unsignedTxsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (n *Node) getTxBytesFromStorage(hash []byte) ([]byte, transaction.TxType, bool) {
	txsStorer := n.store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeNormal, true
	}

	rewardTxsStorer := n.store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeReward, true
	}

	unsignedTxsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (n *Node) getTxBytesFromStorageByEpoch(hash []byte, epoch uint32) ([]byte, transaction.TxType, bool) {
	txsStorer := n.store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeNormal, true
	}

	rewardTxsStorer := n.store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeReward, true
	}

	unsignedTxsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (n *Node) castObjToTransaction(txObj interface{}, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case transaction.TxTypeNormal:
		if tx, ok := txObj.(*transaction.Transaction); ok {
			return n.prepareNormalTx(tx)
		}
	case transaction.TxTypeInvalid:
		if tx, ok := txObj.(*transaction.Transaction); ok {
			return n.prepareInvalidTx(tx)
		}
	case transaction.TxTypeReward:
		if tx, ok := txObj.(*rewardTxData.RewardTx); ok {
			return n.prepareRewardTx(tx)
		}
	case transaction.TxTypeUnsigned:
		if tx, ok := txObj.(*smartContractResult.SmartContractResult); ok {
			return n.prepareUnsignedTx(tx)
		}
	}

	log.Warn("castObjToTransaction() unexpected: unknown txType", "txType", txType)
	return &transaction.ApiTransactionResult{Type: string(transaction.TxTypeInvalid)}, nil
}

func (n *Node) unmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case transaction.TxTypeNormal:
		var tx transaction.Transaction
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return n.prepareNormalTx(&tx)
	case transaction.TxTypeInvalid:
		var tx transaction.Transaction
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return n.prepareInvalidTx(&tx)
	case transaction.TxTypeReward:
		var tx rewardTxData.RewardTx
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return n.prepareRewardTx(&tx)

	case transaction.TxTypeUnsigned:
		var tx smartContractResult.SmartContractResult
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return n.prepareUnsignedTx(&tx)
	}

	return &transaction.ApiTransactionResult{Type: string(transaction.TxTypeInvalid)}, nil // this shouldn't happen
}

func (n *Node) prepareNormalTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeNormal),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         n.addressPubkeyConverter.Encode(tx.RcvAddr),
		ReceiverUsername: tx.RcvUserName,
		Sender:           n.addressPubkeyConverter.Encode(tx.SndAddr),
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
	}, nil
}

func (n *Node) prepareInvalidTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeInvalid),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         n.addressPubkeyConverter.Encode(tx.RcvAddr),
		ReceiverUsername: tx.RcvUserName,
		Sender:           n.addressPubkeyConverter.Encode(tx.SndAddr),
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
	}, nil
}

func (n *Node) prepareRewardTx(tx *rewardTxData.RewardTx) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:          tx,
		Type:        string(transaction.TxTypeReward),
		Round:       tx.GetRound(),
		Epoch:       tx.GetEpoch(),
		Value:       tx.GetValue().String(),
		Sender:      "metachain",
		Receiver:    n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		SourceShard: core.MetachainShardId,
	}, nil
}

func (n *Node) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	txResult := &transaction.ApiTransactionResult{
		Tx:                      tx,
		Type:                    string(transaction.TxTypeUnsigned),
		Nonce:                   tx.GetNonce(),
		Value:                   tx.GetValue().String(),
		Receiver:                n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		Sender:                  n.addressPubkeyConverter.Encode(tx.GetSndAddr()),
		GasPrice:                tx.GetGasPrice(),
		GasLimit:                tx.GetGasLimit(),
		Data:                    tx.GetData(),
		Code:                    string(tx.GetCode()),
		CodeMetadata:            tx.GetCodeMetadata(),
		PreviousTransactionHash: hex.EncodeToString(tx.GetPrevTxHash()),
		OriginalTransactionHash: hex.EncodeToString(tx.GetOriginalTxHash()),
		ReturnMessage:           string(tx.GetReturnMessage()),
	}
	if len(tx.GetOriginalSender()) == n.addressPubkeyConverter.Len() {
		txResult.OriginalSender = n.addressPubkeyConverter.Encode(tx.GetOriginalSender())
	}

	return txResult, nil
}
