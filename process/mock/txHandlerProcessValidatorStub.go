package mock

import "github.com/ElrondNetwork/elrond-go/data"

type TxHandlerProcessValidatorStub struct {
	CheckTxHandlerValidCalled func(txHandler data.TransactionHandler) bool
}

func (t *TxHandlerProcessValidatorStub) CheckTxHandlerValid(txHandler data.TransactionHandler) bool {
	return t.CheckTxHandlerValidCalled(txHandler)
}
