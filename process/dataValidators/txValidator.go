package dataValidators

import (
	"encoding/hex"
	"fmt"
	"math/big"

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
	// TODO: Refactor, extract methods.

	shardID := txv.shardCoordinator.SelfId()
	txShardID := interceptedTx.SenderShardId()
	senderIsInAnotherShard := shardID != txShardID
	if senderIsInAnotherShard {
		return nil
	}

	senderAddress := interceptedTx.SenderAddress()
	accountHandler, err := txv.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return errSenderNotInCurrentShard(senderAddress.Bytes(), shardID)
	}

	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.Nonce()
	lowerNonceInTx := txNonce < accountNonce
	veryHighNonceInTx := txNonce > accountNonce+uint64(txv.maxNonceDeltaAllowed)
	isTxRejected := lowerNonceInTx || veryHighNonceInTx
	if isTxRejected {
		return errInvalidNonce(accountNonce, txNonce)
	}

	account, ok := accountHandler.(*state.Account)
	if !ok {
		return errConvertError(senderAddress.Bytes())
	}

	accountBalance := account.Balance
	txFee := interceptedTx.Fee()
	if accountBalance.Cmp(txFee) < 0 {
		return errInsufficientBalance(txFee, accountBalance)
	}

	return nil
}

func errSenderNotInCurrentShard(senderAddress []byte, currentShard uint32) error {
	return fmt.Errorf("transaction's sender address %s does not exist in current shard %d", hex.EncodeToString(senderAddress), currentShard)
}

func errInvalidNonce(accountNonce uint64, txNonce uint64) error {
	return fmt.Errorf("invalid nonce. Wanted %d, got %d", accountNonce, txNonce)
}

func errConvertError(senderAddress []byte) error {
	return fmt.Errorf("cannot convert account handler in a state.Account %s", hex.EncodeToString(senderAddress))
}

func errInsufficientBalance(txFee *big.Int, accountBalance *big.Int) error {
	return fmt.Errorf("insufficient balance. Needed at least %d (fee), account has %d", txFee, accountBalance)
}

// IsInterfaceNil returns true if there is no value under the interface
func (txv *txValidator) IsInterfaceNil() bool {
	return txv == nil
}
