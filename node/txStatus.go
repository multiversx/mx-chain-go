package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

func (n *Node) GetTransactionStatus(txHash string) (string, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return "", err
	}

	foundInStorage := n.isTxInStorage(hash)
	if foundInStorage {
		return "executed", nil
	}

	foundInDataPool := n.isTxInDataPool(hash)
	if foundInDataPool {
		return "received", nil
	}

	return "unknown", nil
}

func (n *Node) isTxInDataPool(hash []byte) bool {
	// TODO check n.dataPool.UnsignedTransactions() and RewardTransactions()
	txPool := n.dataPool.Transactions()

	_, found := txPool.SearchFirstData(hash)
	return found
}

func (n *Node) isTxInStorage(hash []byte) bool {
	// TODO check UnsignedTransactions and RewardTransactions
	txStorer := n.store.GetStorer(dataRetriever.TransactionUnit)

	err := txStorer.Has(hash)
	return err == nil
}
