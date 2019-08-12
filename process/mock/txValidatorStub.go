package mock

import "github.com/ElrondNetwork/elrond-go/data"

type TxValidatorStub struct {
	IsTxValidForProcessingCalled func(txHandler data.TransactionHandler) bool
}

func (t *TxValidatorStub) IsTxValidForProcessing(txHandler data.TransactionHandler) bool {
	return t.IsTxValidForProcessingCalled(txHandler)
}
