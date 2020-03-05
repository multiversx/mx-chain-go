package dataValidators

import (
	"encoding/hex"
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
		return fmt.Errorf("transaction's sender address %s does not exist in current shard %d",
			hex.EncodeToString(sndAddrBytes),
			shardId)
	}

	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.Nonce()
	lowerNonceInTx := txNonce < accountNonce
	veryHighNonceInTx := txNonce > accountNonce+uint64(txv.maxNonceDeltaAllowed)
	isTxRejected := lowerNonceInTx || veryHighNonceInTx
	if isTxRejected {
		return fmt.Errorf("invalid nonce. Wanted %d, got %d", accountNonce, txNonce)
	}

	account, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		hexSenderAddr := hex.EncodeToString(sndAddr.Bytes())
		return fmt.Errorf("cannot convert account handler in a state.Account %s", hexSenderAddr)
	}

	accountBalance := account.GetBalance()
	txFee := interceptedTx.Fee()
	if accountBalance.Cmp(txFee) < 0 {
		return fmt.Errorf("insufficient balance. Needed at least %d (fee), account has %d", txFee, accountBalance)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txv *txValidator) IsInterfaceNil() bool {
	return txv == nil
}
