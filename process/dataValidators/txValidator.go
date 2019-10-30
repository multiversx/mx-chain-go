package dataValidators

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// txValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type txValidator struct {
	accounts             state.AccountsAdapter
	shardCoordinator     sharding.Coordinator
	rejectedTxs          uint64
	maxNonceDeltaAllowed int
}

// NewTxValidator creates a new nil tx handler validator instance
func NewTxValidator(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	maxNonceDeltaAllowed int,
) (*txValidator, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	return &txValidator{
		accounts:             accounts,
		shardCoordinator:     shardCoordinator,
		rejectedTxs:          uint64(0),
		maxNonceDeltaAllowed: maxNonceDeltaAllowed,
	}, nil
}

// CheckTxValidity will filter transactions that needs to be added in pools
func (tv *txValidator) CheckTxValidity(interceptedTx process.TxValidatorHandler) error {
	shardId := tv.shardCoordinator.SelfId()
	txShardId := interceptedTx.SenderShardId()
	senderIsInAnotherShard := shardId != txShardId
	if senderIsInAnotherShard {
		return nil
	}

	sndAddr := interceptedTx.SenderAddress()
	accountHandler, err := tv.accounts.GetExistingAccount(sndAddr)
	if err != nil {
		sndAddrBytes := sndAddr.Bytes()
		return errors.New(fmt.Sprintf("Transaction's sender address %s does not exist in current shard %d",
			hex.EncodeToString(sndAddrBytes),
			shardId))
	}

	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.Nonce()
	lowerNonceInTx := txNonce < accountNonce
	veryHighNonceInTx := txNonce > accountNonce+uint64(tv.maxNonceDeltaAllowed)
	isTxRejected := lowerNonceInTx || veryHighNonceInTx
	if isTxRejected {
		tv.rejectedTxs++
		return errors.New(fmt.Sprintf("Invalid nonce. Wanted %d, got %d", accountNonce, txNonce))
	}

	account, ok := accountHandler.(*state.Account)
	if !ok {
		hexSenderAddr := hex.EncodeToString(sndAddr.Bytes())
		return errors.New(fmt.Sprintf("Cannot convert account handler in a state.Account %s", hexSenderAddr))
	}

	accountBalance := account.Balance
	txTotalValue := interceptedTx.TotalValue()
	if accountBalance.Cmp(txTotalValue) < 0 {
		tv.rejectedTxs++
		return errors.New(fmt.Sprintf("Insufficient balance. Needed %d ERD, account has %d ERD", txTotalValue, accountBalance))
	}

	return nil
}

// NumRejectedTxs will return number of rejected transaction
func (tv *txValidator) NumRejectedTxs() uint64 {
	return tv.rejectedTxs
}

// IsInterfaceNil returns true if there is no value under the interface
func (tv *txValidator) IsInterfaceNil() bool {
	if tv == nil {
		return true
	}
	return false
}
