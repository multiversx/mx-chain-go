package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type TxValidatorStub struct {
	CheckTxValidityCalled func(txValidatorHandler process.TxValidatorHandler) error
	RejectedTxsCalled     func() uint64
}

func (t *TxValidatorStub) CheckTxValidity(txValidatorHandler process.TxValidatorHandler) error {
	return t.CheckTxValidityCalled(txValidatorHandler)
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
