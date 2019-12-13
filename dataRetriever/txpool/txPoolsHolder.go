package txpool

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

import "github.com/ElrondNetwork/elrond-go/core/check"

// TxPoolsHolder is a helper structure that holds transactions pools for: regular transactions, unsigned transactions and reward transactions
type TxPoolsHolder struct {
	transactions         dataRetriever.TxPool
	unsignedTransactions dataRetriever.TxPool
	rewardTransactions   dataRetriever.TxPool
}

// NewTxPoolsHolder creates a group of transaction pools
func NewTxPoolsHolder(transactions dataRetriever.TxPool, unsignedTransactions dataRetriever.TxPool, rewardTransactions dataRetriever.TxPool) *TxPoolsHolder {
	return &TxPoolsHolder{
		transactions:         transactions,
		unsignedTransactions: unsignedTransactions,
		rewardTransactions:   rewardTransactions,
	}
}

// TransactionsPool gets the transaction pool from the group
func (group *TxPoolsHolder) TransactionsPool() dataRetriever.TxPool {
	check.AssertNotNil(group.transactions, "transactions pool")
	return group.transactions
}

// UnsignedTransactions gets the unsigned transaction pool from the group
func (group *TxPoolsHolder) UnsignedTransactions() dataRetriever.TxPool {
	check.AssertNotNil(group.unsignedTransactions, "unsigned transactions pool")
	return group.unsignedTransactions
}

// RewardTransactions gets the unsigned transaction pool from the group
func (group *TxPoolsHolder) RewardTransactions() dataRetriever.TxPool {
	check.AssertNotNil(group.rewardTransactions, "reward transactions pool")
	return group.rewardTransactions
}

// IsInterfaceNil returns true if there is no value under the interface
func (group *TxPoolsHolder) IsInterfaceNil() bool {
	if group == nil {
		return true
	}
	return false
}
