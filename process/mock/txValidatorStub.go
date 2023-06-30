package mock

import (
	"github.com/multiversx/mx-chain-go/process"
)

// TxValidatorStub -
type TxValidatorStub struct {
	CheckTxValidityCalled   func(interceptedTx process.InterceptedTransactionHandler) error
	CheckTxWhiteListCalled  func(data process.InterceptedData) error
	RejectedTxsCalled       func() uint64
}

// CheckTxValidity -
func (t *TxValidatorStub) CheckTxValidity(interceptedTx process.InterceptedTransactionHandler) error {
	return t.CheckTxValidityCalled(interceptedTx)
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
