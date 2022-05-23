package dataValidators

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// CheckAccount -
func (txv *txValidator) CheckAccount(
	interceptedTx process.InterceptedTransactionHandler,
	accountHandler vmcommon.AccountHandler,
) error {
	return txv.checkAccount(interceptedTx, accountHandler)
}

// CheckPermission -
func (txv *txValidator) CheckPermission(
	interceptedTx process.InterceptedTransactionHandler,
	account state.UserAccountHandler,
) error {
	return txv.checkPermission(interceptedTx, account)
}

// CheckGuardedTransaction -
func (txv *txValidator) CheckGuardedTransaction(
	interceptedTx process.InterceptedTransactionHandler,
	account state.UserAccountHandler,
) error {
	return txv.checkGuardedTransaction(interceptedTx, account)
}

// CheckOperationAllowedToBypassGuardian -
func CheckOperationAllowedToBypassGuardian(txData []byte) error {
	return checkOperationAllowedToBypassGuardian(txData)
}

// GetTxData -
func GetTxData(interceptedTx process.InterceptedTransactionHandler) ([]byte, error) {
	return getTxData(interceptedTx)
}

// IsBuiltInFuncCallWithParam -
func IsBuiltInFuncCallWithParam(txData []byte, function string) bool {
	return isBuiltinFuncCallWithParam(txData, function)
}
