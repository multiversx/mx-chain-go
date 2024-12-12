package txcachemocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// MempoolHostMock -
type MempoolHostMock struct {
	ComputeTxFeeCalled        func(tx data.TransactionWithFeeHandler) *big.Int
	GetTransferredValueCalled func(tx data.TransactionHandler) *big.Int
}

// NewMempoolHostMock -
func NewMempoolHostMock() *MempoolHostMock {
	return &MempoolHostMock{}
}

// ComputeTxFee -
func (mock *MempoolHostMock) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	if mock.ComputeTxFeeCalled != nil {
		return mock.ComputeTxFeeCalled(tx)
	}

	return big.NewInt(0)
}

// GetTransferredValue -
func (mock *MempoolHostMock) GetTransferredValue(tx data.TransactionHandler) *big.Int {
	if mock.GetTransferredValueCalled != nil {
		return mock.GetTransferredValueCalled(tx)
	}

	return tx.GetValue()
}

// IsInterfaceNil -
func (mock *MempoolHostMock) IsInterfaceNil() bool {
	return mock == nil
}
