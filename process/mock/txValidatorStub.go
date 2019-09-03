package mock

import "github.com/ElrondNetwork/elrond-go/data"

type TxValidatorStub struct {
	IsTxValidForProcessingCalled func(txHandler data.TransactionHandler) bool
}

func (t *TxValidatorStub) IsTxValidForProcessing(txHandler data.TransactionHandler) bool {
	return t.IsTxValidForProcessingCalled(txHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *TxValidatorStub) IsInterfaceNil() bool {
	if t == nil {
		return true
	}
	return false
}
