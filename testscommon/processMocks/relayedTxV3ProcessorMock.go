package processMocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// RelayedTxV3ProcessorMock -
type RelayedTxV3ProcessorMock struct {
	ComputeRelayedTxFeesCalled func(tx *transaction.Transaction) (*big.Int, *big.Int)
	CheckRelayedTxCalled       func(tx *transaction.Transaction) error
}

// ComputeRelayedTxFees -
func (mock *RelayedTxV3ProcessorMock) ComputeRelayedTxFees(tx *transaction.Transaction) (*big.Int, *big.Int) {
	if mock.ComputeRelayedTxFeesCalled != nil {
		return mock.ComputeRelayedTxFeesCalled(tx)
	}
	return nil, nil
}

// CheckRelayedTx -
func (mock *RelayedTxV3ProcessorMock) CheckRelayedTx(tx *transaction.Transaction) error {
	if mock.CheckRelayedTxCalled != nil {
		return mock.CheckRelayedTxCalled(tx)
	}
	return nil
}

// IsInterfaceNil -
func (mock *RelayedTxV3ProcessorMock) IsInterfaceNil() bool {
	return mock == nil
}
