package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type TxValidatorStub struct {
	IsTxValidForProcessingCalled func(txValidatorHandler process.TxValidatorHandler) bool
	RejectedTxsCalled            func() uint64
}

func (t *TxValidatorStub) IsTxValidForProcessing(txValidatorHandler process.TxValidatorHandler) bool {
	return t.IsTxValidForProcessingCalled(txValidatorHandler)
}

func (t *TxValidatorStub) NumRejectedTxs() uint64 {
	return t.RejectedTxsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *TxValidatorStub) IsInterfaceNil() bool {
	if t == nil {
		return true
	}
	return false
}
