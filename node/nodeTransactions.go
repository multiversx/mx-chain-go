package node

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	rewardTxData "github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type transactionType string

const (
	normalTx   transactionType = "normal"
	unsignedTx transactionType = "unsignedTx"
	rewardTx   transactionType = "rewardTx"
	invalidTx  transactionType = "invalidTx"
)

// GetTransaction gets the transaction based on the given hash. It will search in the cache and the storage and
// will return the transaction in a format which can be respected by all types of transactions (normal, reward or unsigned)
func (n *Node) GetTransaction(txHash string) (*transaction.ApiTransactionResult, error) {
	if !n.apiTransactionByHashThrottler.CanProcess() {
		return nil, ErrSystemBusyTxHash
	}

	n.apiTransactionByHashThrottler.StartProcessing()
	defer n.apiTransactionByHashThrottler.EndProcessing()

	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	txBytes, txType, found := n.getTxBytesFromDataPool(hash)
	if found {
		return n.unmarshalTransaction(txBytes, txType)
	}

	txBytes, txType, found = n.getTxBytesFromStorage(hash)
	if found {
		return n.unmarshalTransaction(txBytes, txType)
	}

	return nil, fmt.Errorf("transaction not found")
}

// GetTransactionStatus gets the transaction status
func (n *Node) GetTransactionStatus(txHash string) (string, error) {
	if !n.apiTransactionByHashThrottler.CanProcess() {
		return "", ErrSystemBusyTxHash
	}

	n.apiTransactionByHashThrottler.StartProcessing()
	defer n.apiTransactionByHashThrottler.EndProcessing()

	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return "", err
	}

	_, _, foundInDataPool := n.getTxBytesFromDataPool(hash)
	if foundInDataPool {
		return string(core.TxStatusReceived), nil
	}

	foundInStorage := n.isTxInStorage(hash)
	if foundInStorage {
		return string(core.TxStatusExecuted), nil
	}

	return string(core.TxStatusUnknown), nil
}

func (n *Node) getTxBytesFromDataPool(hash []byte) ([]byte, transactionType, bool) {
	txsPool := n.dataPool.Transactions()
	txBytes, found := txsPool.SearchFirstData(hash)
	if found && txBytes != nil {
		return txBytes.([]byte), normalTx, true
	}

	rewardTxsPool := n.dataPool.RewardTransactions()
	txBytes, found = rewardTxsPool.SearchFirstData(hash)
	if found && txBytes != nil {
		return txBytes.([]byte), rewardTx, true
	}

	unsignedTxsPool := n.dataPool.UnsignedTransactions()
	txBytes, found = unsignedTxsPool.SearchFirstData(hash)
	if found && txBytes != nil {
		return txBytes.([]byte), unsignedTx, true
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

func (n *Node) unmarshalTransaction(txBytes []byte, txType transactionType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case normalTx:
		var tx transaction.Transaction
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return &transaction.ApiTransactionResult{
			Type:      string(normalTx),
			Nonce:     tx.Nonce,
			Value:     tx.Value.String(),
			Receiver:  n.addressPubkeyConverter.Encode(tx.RcvAddr),
			Sender:    n.addressPubkeyConverter.Encode(tx.SndAddr),
			GasPrice:  tx.GasPrice,
			GasLimit:  tx.GasLimit,
			Data:      string(tx.Data),
			Signature: hex.EncodeToString(tx.Signature),
		}, nil
	case rewardTx:
		var tx rewardTxData.RewardTx
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return &transaction.ApiTransactionResult{
			Type:     string(rewardTx),
			Round:    tx.GetRound(),
			Epoch:    tx.GetEpoch(),
			Value:    tx.GetValue().String(),
			Receiver: n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
		}, nil

	case unsignedTx:
		var tx smartContractResult.SmartContractResult
		err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return &transaction.ApiTransactionResult{
			Type:      string(unsignedTx),
			Nonce:     tx.GetNonce(),
			Value:     tx.GetValue().String(),
			Receiver:  n.addressPubkeyConverter.Encode(tx.GetRcvAddr()),
			Sender:    n.addressPubkeyConverter.Encode(tx.GetSndAddr()),
			GasPrice:  tx.GetGasPrice(),
			GasLimit:  tx.GetGasLimit(),
			Data:      string(tx.GetData()),
			Code:      string(tx.GetCode()),
			Signature: "",
		}, nil
	default:
		return &transaction.ApiTransactionResult{Type: string(invalidTx)}, nil // this shouldn't happen
	}
}
