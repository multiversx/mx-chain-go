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
		maxNonceDeltaAllowed: maxNonceDeltaAllowed,
	}, nil
}

// CheckTxValidity will filter transactions that needs to be added in pools
func (txv *txValidator) CheckTxValidity(interceptedTx process.TxValidatorHandler) error {
	shardId := txv.shardCoordinator.SelfId()
	txShardId := interceptedTx.SenderShardId()
	senderIsInAnotherShard := shardId != txShardId
	if senderIsInAnotherShard {
		return nil
	}

	sndAddr := interceptedTx.SenderAddress()
	accountHandler, err := txv.accounts.GetExistingAccount(sndAddr)
	if err != nil {
		sndAddrBytes := sndAddr.Bytes()
		return errors.New(fmt.Sprintf("transaction's sender address %s does not exist in current shard %d",
			hex.EncodeToString(sndAddrBytes),
			shardId))
	}

	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.Nonce()
	lowerNonceInTx := txNonce < accountNonce
	veryHighNonceInTx := txNonce > accountNonce+uint64(txv.maxNonceDeltaAllowed)
	isTxRejected := lowerNonceInTx || veryHighNonceInTx
	if isTxRejected {
		return errors.New(fmt.Sprintf("Invalid nonce. Wanted %d, got %d", accountNonce, txNonce))
	}

	account, ok := accountHandler.(*state.Account)
	if !ok {
		hexSenderAddr := hex.EncodeToString(sndAddr.Bytes())
		return errors.New(fmt.Sprintf("cannot convert account handler in a state.Account %s", hexSenderAddr))
	}

	txTotalValue := interceptedTx.TotalValue()
	if account.Balance.Cmp(txTotalValue) < 0 {
		return errors.New(fmt.Sprintf("insufficient balance. Needed %d, account has %d", txTotalValue, account.Balance))
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txv *txValidator) IsInterfaceNil() bool {
	if txv == nil {
		return true
	}
	return false
}
