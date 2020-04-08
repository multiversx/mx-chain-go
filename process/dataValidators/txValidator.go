package dataValidators

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// txValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type txValidator struct {
	accounts             state.AccountsAdapter
	shardCoordinator     sharding.Coordinator
	pubkeyConverter      state.PubkeyConverter
	maxNonceDeltaAllowed int
}

// NewTxValidator creates a new nil tx handler validator instance
func NewTxValidator(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	maxNonceDeltaAllowed int,
	pubkeyConverter state.PubkeyConverter,
) (*txValidator, error) {

	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(pubkeyConverter) {
		return nil, fmt.Errorf("%w in NewTxValidator", process.ErrNilPubkeyConverter)
	}

	return &txValidator{
		accounts:             accounts,
		shardCoordinator:     shardCoordinator,
		maxNonceDeltaAllowed: maxNonceDeltaAllowed,
		pubkeyConverter:      pubkeyConverter,
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
		return fmt.Errorf("%w for address %s and shard %d",
			process.ErrAddressNotInThisShard,
			txv.pubkeyConverter.Encode(senderAddress.Bytes()),
			shardID,
		)
	}

	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.Nonce()
	lowerNonceInTx := txNonce < accountNonce
	veryHighNonceInTx := txNonce > accountNonce+uint64(txv.maxNonceDeltaAllowed)
	isTxRejected := lowerNonceInTx || veryHighNonceInTx
	if isTxRejected {
		return fmt.Errorf("%w lowerNonceInTx: %v, veryHighNonceInTx: %v",
			process.ErrWrongTransaction,
			lowerNonceInTx,
			veryHighNonceInTx,
		)
	}

	account, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return fmt.Errorf("%w, account is not of type *state.Account, address: %s",
			process.ErrWrongTypeAssertion,
			txv.pubkeyConverter.Encode(senderAddress.Bytes()),
		)
	}

	accountBalance := account.GetBalance()
	txFee := interceptedTx.Fee()
	if accountBalance.Cmp(txFee) < 0 {
		return fmt.Errorf("%w, for address: %s, wanted %v, have %v",
			process.ErrInsufficientFunds,
			txv.pubkeyConverter.Encode(senderAddress.Bytes()),
			txFee,
			accountBalance,
		)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txv *txValidator) IsInterfaceNil() bool {
	return txv == nil
}
