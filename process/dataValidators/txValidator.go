package dataValidators

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.TxValidator = (*txValidator)(nil)

// GuardianSigVerifier allows the verification of the guardian signatures for guarded transactions
// TODO: add an implementation and integrate it
type GuardianSigVerifier interface {
	VerifyGuardianSignature(guardianPubKey []byte, tx data.TransactionHandler) error // todo: through crypto.SingleSigner
	GetActiveGuardian(handler data.UserAccountHandler) ([]byte, error)               // todo: through core.GuardianChecker
	IsInterfaceNil() bool
}

// txValidator represents a tx handler validator that doesn't check the validity of provided txHandler
type txValidator struct {
	accounts             state.AccountsAdapter
	shardCoordinator     sharding.Coordinator
	whiteListHandler     process.WhiteListHandler
	pubKeyConverter      core.PubkeyConverter
	guardianSigVerifier  GuardianSigVerifier
	maxNonceDeltaAllowed int
}

// NewTxValidator creates a new nil tx handler validator instance
func NewTxValidator(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	whiteListHandler process.WhiteListHandler,
	pubKeyConverter core.PubkeyConverter,
	guardianSigVerifier GuardianSigVerifier,
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

	return &txValidator{
		accounts:             accounts,
		shardCoordinator:     shardCoordinator,
		whiteListHandler:     whiteListHandler,
		maxNonceDeltaAllowed: maxNonceDeltaAllowed,
		pubKeyConverter:      pubKeyConverter,
		guardianSigVerifier:  guardianSigVerifier,
	}, nil
}

// CheckTxValidity will filter transactions that needs to be added in pools
func (txv *txValidator) CheckTxValidity(interceptedTx process.InterceptedTxHandler) error {
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
	interceptedTx process.InterceptedTxHandler,
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
	interceptedTx process.InterceptedTxHandler,
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

func (txv *txValidator) checkBalance(interceptedTx process.InterceptedTxHandler, account state.UserAccountHandler) error {
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

func (txv *txValidator) checkPermission(interceptedTx process.InterceptedTxHandler, account state.UserAccountHandler) error {
	txData, err := getTxData(interceptedTx)
	if err != nil {
		return err
	}
	if isAccountFrozen(account) {
		if isBuiltinFuncCallWithParam(txData, core.BuiltInFunctionSetGuardian) {
			return state.ErrOperationNotPermitted
		}

		guardianPubKey, errGuardian := txv.guardianSigVerifier.GetActiveGuardian(account)
		if errGuardian != nil {
			return fmt.Errorf("%w due to error in getting the active guardian %s", state.ErrOperationNotPermitted, errGuardian.Error())
		}

		errGuardianSignature := txv.guardianSigVerifier.VerifyGuardianSignature(guardianPubKey, interceptedTx.Transaction())
		if errGuardianSignature != nil {
			return fmt.Errorf("%w due to error in signature verification %s", state.ErrOperationNotPermitted, errGuardianSignature.Error())
		}
	}

	return nil
}

func (txv *txValidator) checkNonce(interceptedTx process.InterceptedTxHandler, accountHandler vmcommon.AccountHandler) error {
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

func (txv *txValidator) isSenderInDifferentShard(interceptedTx process.InterceptedTxHandler) bool {
	shardID := txv.shardCoordinator.SelfId()
	txShardID := interceptedTx.SenderShardId()
	return shardID != txShardID
}

func (txv *txValidator) getSenderAccount(interceptedTx process.InterceptedTxHandler) (vmcommon.AccountHandler, error) {
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

func getTxData(interceptedTx process.InterceptedTxHandler) ([]byte, error) {
	tx := interceptedTx.Transaction()
	if tx == nil {
		return nil, process.ErrNilTransaction
	}

	return tx.GetData(), nil
}

func isAccountFrozen(account state.UserAccountHandler) bool {
	codeMetaDataBytes := account.GetCodeMetadata()
	codeMetaData := vmcommon.CodeMetadataFromBytes(codeMetaDataBytes)
	return codeMetaData.Frozen
}

func isBuiltinFuncCallWithParam(txData []byte, function string) bool {
	expectedTxDataPrefix := []byte(function + "@")
	return bytes.HasPrefix(txData, expectedTxDataPrefix)
}

// CheckTxWhiteList will check if the cross shard transactions are whitelisted and could be added in pools
func (txv *txValidator) CheckTxWhiteList(data process.InterceptedData) error {
	interceptedTx, ok := data.(processor.InterceptedTransactionHandler)
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
