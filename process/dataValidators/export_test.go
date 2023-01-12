package dataValidators

import (
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// CheckAccount -
func (txv *txValidator) CheckAccount(
	interceptedTx process.InterceptedTransactionHandler,
	accountHandler vmcommon.AccountHandler,
) error {
	return txv.checkAccount(interceptedTx, accountHandler)
}

// GetTxData -
func GetTxData(interceptedTx process.InterceptedTransactionHandler) ([]byte, error) {
	return getTxData(interceptedTx)
}
