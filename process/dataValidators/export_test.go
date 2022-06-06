package dataValidators

import "github.com/ElrondNetwork/elrond-go/process"

func (txv *txValidator) CheckBlacklist(interceptedTx process.TxValidatorHandler) error {
	return txv.checkBlacklist(interceptedTx)
}