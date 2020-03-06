package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// UnsignedTxHandlerMock -
type UnsignedTxHandlerMock struct {
	CleanProcessedUtxsCalled    func()
	ProcessTransactionFeeCalled func(cost *big.Int)
	CreateAllUTxsCalled         func() []data.TransactionHandler
	VerifyCreatedUTxsCalled     func() error
	AddTxFeeFromBlockCalled     func(tx data.TransactionHandler)
	GetAccumulatedFeesCalled    func() *big.Int
}

// CreateBlockStarted -
func (ut *UnsignedTxHandlerMock) CreateBlockStarted() {
}

// GetAccumulatedFees -
func (ut *UnsignedTxHandlerMock) GetAccumulatedFees() *big.Int {
	if ut.GetAccumulatedFeesCalled != nil {
		return ut.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

// AddRewardTxFromBlock -
func (ut *UnsignedTxHandlerMock) AddRewardTxFromBlock(tx data.TransactionHandler) {
	if ut.AddTxFeeFromBlockCalled == nil {
		return
	}

	ut.AddTxFeeFromBlockCalled(tx)
}

// CleanProcessedUTxs -
func (ut *UnsignedTxHandlerMock) CleanProcessedUTxs() {
	if ut.CleanProcessedUtxsCalled == nil {
		return
	}

	ut.CleanProcessedUtxsCalled()
}

// ProcessTransactionFee -
func (ut *UnsignedTxHandlerMock) ProcessTransactionFee(cost *big.Int) {
	if ut.ProcessTransactionFeeCalled == nil {
		return
	}

	ut.ProcessTransactionFeeCalled(cost)
}

// CreateAllUTxs -
func (ut *UnsignedTxHandlerMock) CreateAllUTxs() []data.TransactionHandler {
	if ut.CreateAllUTxsCalled == nil {
		return nil
	}
	return ut.CreateAllUTxsCalled()
}

// VerifyCreatedUTxs -
func (ut *UnsignedTxHandlerMock) VerifyCreatedUTxs() error {
	if ut.VerifyCreatedUTxsCalled == nil {
		return nil
	}
	return ut.VerifyCreatedUTxsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ut *UnsignedTxHandlerMock) IsInterfaceNil() bool {
	return ut == nil
}
