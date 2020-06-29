package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// TxValidatorStub -
type TxValidatorStub struct {
	CheckTxValidityCalled  func(txValidatorHandler process.TxValidatorHandler) error
	CheckTxWhiteListCalled func(data process.InterceptedData) error
	RejectedTxsCalled      func() uint64
}

// CheckTxValidity -
func (t *TxValidatorStub) CheckTxValidity(txValidatorHandler process.TxValidatorHandler) error {
	return t.CheckTxValidityCalled(txValidatorHandler)
}

// CheckTxWhiteList -
func (t *TxValidatorStub) CheckTxWhiteList(data process.InterceptedData) error {
	if t.CheckTxWhiteListCalled != nil {
		return t.CheckTxWhiteListCalled(data)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *TxValidatorStub) IsInterfaceNil() bool {
	return t == nil
}
