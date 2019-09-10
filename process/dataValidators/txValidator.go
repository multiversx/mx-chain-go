package dataValidators

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/prometheus/common/log"
)

// TxValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type TxValidator struct {
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
	rejectedTxs      uint64
}

// NewTxValidator creates a new nil tx handler validator instance
func NewTxValidator(accounts state.AccountsAdapter, shardCoordinator sharding.Coordinator) (*TxValidator, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	return &TxValidator{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		rejectedTxs:      uint64(0),
	}, nil
}

// IsTxValidForProcessing will filter transactions that needs to be added in pools
func (tv *TxValidator) IsTxValidForProcessing(interceptedTx process.TxValidatorHandler) bool {
	shardId := tv.shardCoordinator.SelfId()
	txShardId := interceptedTx.GetSenderShardId()
	if shardId != txShardId {
		return true
	}

	sndAddr := interceptedTx.GetSenderAddress()
	accountHandler, err := tv.accounts.GetExistingAccount(sndAddr)
	if err != nil {
		log.Debug(fmt.Sprintf("Transaction sender address %s does not exit in current shard %d", sndAddr, shardId))
		tv.rejectedTxs++
		return false
	}

	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.GetNonce()
	if txNonce < accountNonce {
		tv.rejectedTxs++
		return false
	}

	account, ok := accountHandler.(*state.Account)
	if !ok {
		log.Error("Cannot convert account handler in a state.Account")
		return false
	}

	accountBalance := account.Balance
	txTotalValue := interceptedTx.GetTotalValue()
	if accountBalance.Cmp(txTotalValue) < 0 {
		tv.rejectedTxs++
		return false
	}

	return true
}

// GetNumRejectedTxs will return number of rejected transaction
func (tv *TxValidator) GetNumRejectedTxs() uint64 {
	return tv.rejectedTxs
}

// IsInterfaceNil returns true if there is no value under the interface
func (tv *TxValidator) IsInterfaceNil() bool {
	if tv == nil {
		return true
	}
	return false
}
