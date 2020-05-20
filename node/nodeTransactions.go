package node

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// GetTransaction gets the transaction
func (n *Node) GetTransaction(txHash string) (*transaction.Transaction, error) {
	if !n.apiTransactionByHashThrottler.CanProcess() {
		return nil, ErrSystemBusyTxHash
	}

	n.apiTransactionByHashThrottler.StartProcessing()
	defer n.apiTransactionByHashThrottler.EndProcessing()

	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	txBytes, found := n.getTxFromDataPool(hash)
	if found {
		return n.convertBytesToTransaction(txBytes)
	}

	txBytes, found = n.getTxFromStorage(hash)
	if found {
		return n.convertBytesToTransaction(txBytes)
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

	_, foundInDataPool := n.getTxFromDataPool(hash)
	if foundInDataPool {
		return "received", nil
	}

	foundInStorage := n.isTxInStorage(hash)
	if foundInStorage {
		return "executed", nil
	}

	return "unknown", nil
}

func (n *Node) getTxFromDataPool(hash []byte) ([]byte, bool) {
	txsPool := n.dataPool.Transactions()
	txBytes, found := txsPool.SearchFirstData(hash)
	if found && txBytes != nil {
		return txBytes.([]byte), true
	}

	rewardTxsPool := n.dataPool.RewardTransactions()
	txBytes, found = rewardTxsPool.SearchFirstData(hash)
	if found && txBytes != nil {
		return txBytes.([]byte), true
	}

	unsignedTxsPool := n.dataPool.UnsignedTransactions()
	txBytes, found = unsignedTxsPool.SearchFirstData(hash)
	if found && txBytes != nil {
		return txBytes.([]byte), true
	}

	return nil, false
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

	unsignedTransactionsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	err = unsignedTransactionsStorer.Has(hash)
	return err == nil
}

func (n *Node) getTxFromStorage(hash []byte) ([]byte, bool) {
	txsStorer := n.store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, true
	}

	rewardTxsStorer := n.store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, true
	}

	unsignedTransactionsStorer := n.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTransactionsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, true
	}

	return nil, false
}

func (n *Node) convertBytesToTransaction(txBytes []byte) (*transaction.Transaction, error) {
	var tx transaction.Transaction
	err := n.internalMarshalizer.Unmarshal(&tx, txBytes)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}
