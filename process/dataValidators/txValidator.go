package dataValidators

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.TxValidator = (*txValidator)(nil)

// txValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type txValidator struct {
	accounts             state.AccountsAdapter
	shardCoordinator     sharding.Coordinator
	whiteListHandler     process.WhiteListHandler
	pubKeyConverter      core.PubkeyConverter
	guardianSigVerifier  process.GuardianSigVerifier
	txVersionChecker     process.TxVersionCheckerHandler
	maxNonceDeltaAllowed int
}

// NewTxValidator creates a new nil tx handler validator instance
func NewTxValidator(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	whiteListHandler process.WhiteListHandler,
	pubKeyConverter core.PubkeyConverter,
	guardianSigVerifier process.GuardianSigVerifier,
	txVersionChecker process.TxVersionCheckerHandler,
	maxNonceDeltaAllowed int,
) (*txValidator, error) {
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(whiteListHandler) {
		return nil, process.ErrNilWhiteListHandler
	}
	if check.IfNil(pubKeyConverter) {
		return nil, fmt.Errorf("%w in NewTxValidator", process.ErrNilPubkeyConverter)
	}
	if check.IfNil(guardianSigVerifier) {
		return nil, process.ErrNilGuardianSigVerifier
	}
	if check.IfNil(txVersionChecker) {
		return nil, process.ErrNilTransactionVersionChecker
	}

	return &txValidator{
		accounts:             accounts,
		shardCoordinator:     shardCoordinator,
		whiteListHandler:     whiteListHandler,
		maxNonceDeltaAllowed: maxNonceDeltaAllowed,
		pubKeyConverter:      pubKeyConverter,
		guardianSigVerifier:  guardianSigVerifier,
		txVersionChecker:     txVersionChecker,
	}, nil
}

// CheckTxValidity will filter transactions that needs to be added in pools
func (txv *txValidator) CheckTxValidity(interceptedTx process.InterceptedTransactionHandler) error {
	interceptedData, ok := interceptedTx.(process.InterceptedData)
	if ok {
		if txv.whiteListHandler.IsWhiteListed(interceptedData) {
			return nil
		}
	}

	if txv.isSenderInDifferentShard(interceptedTx) {
		return nil
	}

	accountHandler, err := txv.getSenderAccount(interceptedTx)
	if err != nil {
		return err
	}

	return txv.checkAccount(interceptedTx, accountHandler)
}

func (txv *txValidator) checkAccount(
	interceptedTx process.InterceptedTransactionHandler,
	accountHandler vmcommon.AccountHandler,
) error {
	err := txv.checkNonce(interceptedTx, accountHandler)
	if err != nil {
		return err
	}

	account, err := txv.getSenderUserAccount(interceptedTx, accountHandler)
	if err != nil {
		return err
	}

	err = txv.checkPermission(interceptedTx, account)
	if err != nil {
		return err
	}

	return txv.checkBalance(interceptedTx, account)
}

func (txv *txValidator) getSenderUserAccount(
	interceptedTx process.InterceptedTransactionHandler,
	accountHandler vmcommon.AccountHandler,
) (state.UserAccountHandler, error) {
	senderAddress := interceptedTx.SenderAddress()
	account, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("%w, account is not of type *state.Account, address: %s",
			process.ErrWrongTypeAssertion,
			txv.pubKeyConverter.Encode(senderAddress),
		)
	}
	return account, nil
}

func (txv *txValidator) checkBalance(interceptedTx process.InterceptedTransactionHandler, account state.UserAccountHandler) error {
	accountBalance := account.GetBalance()
	txFee := interceptedTx.Fee()
	if accountBalance.Cmp(txFee) < 0 {
		senderAddress := interceptedTx.SenderAddress()
		return fmt.Errorf("%w, for address: %s, wanted %v, have %v",
			process.ErrInsufficientFunds,
			txv.pubKeyConverter.Encode(senderAddress),
			txFee,
			accountBalance,
		)
	}

	return nil
}

func (txv *txValidator) checkPermission(interceptedTx process.InterceptedTransactionHandler, account state.UserAccountHandler) error {
	txData, err := getTxData(interceptedTx)
	if err != nil {
		return err
	}

	if account.IsGuarded() {
		err = txv.checkGuardedTransaction(interceptedTx, account)
		if err == nil {
			return nil
		}

		err = checkOperationAllowedToBypassGuardian(txData)
		if err != nil {
			return err
		}

		// block non guarded setGuardian Txs if there is a pending guardian
		hasPendingGuardian := txv.guardianSigVerifier.HasPendingGuardian(account)
		if process.IsSetGuardianCall(txData) && hasPendingGuardian {
			return process.ErrCannotReplaceGuardedAccountPendingGuardian
		}
	}

	return nil
}

// Setting a guardian is allowed with regular transactions on a guarded account
// but in this case is set with the default epochs delay
func checkOperationAllowedToBypassGuardian(txData []byte) error {
	if process.IsSetGuardianCall(txData) {
		return nil
	}

	return process.ErrOperationNotPermitted
}

func (txv *txValidator) checkGuardedTransaction(interceptedTx process.InterceptedTransactionHandler, account state.UserAccountHandler) error {
	txHandler := interceptedTx.Transaction()
	if check.IfNil(txHandler) {
		return process.ErrNilTransaction
	}

	tx, ok := txHandler.(*transaction.Transaction)
	if !ok {
		return fmt.Errorf("%w on transaction handler", process.ErrWrongTypeAssertion)
	}

	if !txv.txVersionChecker.IsGuardedTransaction(tx) {
		return fmt.Errorf("%w without guardian signature", process.ErrOperationNotPermitted)
	}

	vmUserAccount, ok := account.(vmcommon.UserAccountHandler)
	if !ok {
		return fmt.Errorf("%w on account", process.ErrWrongTypeAssertion)
	}

	errGuardianSignature := txv.guardianSigVerifier.VerifyGuardianSignature(vmUserAccount, interceptedTx)
	if errGuardianSignature != nil {
		return fmt.Errorf("%w due to error in signature verification %s", process.ErrOperationNotPermitted, errGuardianSignature.Error())
	}

	return nil
}

func (txv *txValidator) checkNonce(interceptedTx process.InterceptedTransactionHandler, accountHandler vmcommon.AccountHandler) error {
	accountNonce := accountHandler.GetNonce()
	txNonce := interceptedTx.Nonce()
	lowerNonceInTx := txNonce < accountNonce
	veryHighNonceInTx := txNonce > accountNonce+uint64(txv.maxNonceDeltaAllowed)
	if lowerNonceInTx || veryHighNonceInTx {
		return fmt.Errorf("%w lowerNonceInTx: %v, veryHighNonceInTx: %v",
			process.ErrWrongTransaction,
			lowerNonceInTx,
			veryHighNonceInTx,
		)
	}
	return nil
}

func (txv *txValidator) isSenderInDifferentShard(interceptedTx process.InterceptedTransactionHandler) bool {
	shardID := txv.shardCoordinator.SelfId()
	txShardID := interceptedTx.SenderShardId()
	return shardID != txShardID
}

func (txv *txValidator) getSenderAccount(interceptedTx process.InterceptedTransactionHandler) (vmcommon.AccountHandler, error) {
	senderAddress := interceptedTx.SenderAddress()
	accountHandler, err := txv.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for address %s and shard %d, err: %s",
			process.ErrAccountNotFound,
			txv.pubKeyConverter.Encode(senderAddress),
			txv.shardCoordinator.SelfId(),
			err.Error(),
		)
	}

	return accountHandler, nil
}

func getTxData(interceptedTx process.InterceptedTransactionHandler) ([]byte, error) {
	tx := interceptedTx.Transaction()
	if tx == nil {
		return nil, process.ErrNilTransaction
	}

	return tx.GetData(), nil
}

// CheckTxWhiteList will check if the cross shard transactions are whitelisted and could be added in pools
func (txv *txValidator) CheckTxWhiteList(data process.InterceptedData) error {
	interceptedTx, ok := data.(process.InterceptedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	isTxCrossShardDestMe := interceptedTx.SenderShardId() != txv.shardCoordinator.SelfId() &&
		interceptedTx.ReceiverShardId() == txv.shardCoordinator.SelfId()
	if !isTxCrossShardDestMe {
		return nil
	}

	if txv.whiteListHandler.IsWhiteListed(data) {
		return nil
	}

	return process.ErrTransactionIsNotWhitelisted
}

// IsInterfaceNil returns true if there is no value under the interface
func (txv *txValidator) IsInterfaceNil() bool {
	return txv == nil
}
