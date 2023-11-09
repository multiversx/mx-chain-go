package testscommon

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
)

// UnsignedTxHandlerStub -
type UnsignedTxHandlerStub struct {
	CleanProcessedUtxsCalled                 func()
	ProcessTransactionFeeCalled              func(cost *big.Int, fee *big.Int, hash []byte)
	ProcessTransactionFeeRelayedUserTxCalled func(cost *big.Int, devFee *big.Int, userTxHash []byte, originalTxHash []byte)
	CreateAllUTxsCalled                      func() []data.TransactionHandler
	VerifyCreatedUTxsCalled                  func() error
	AddTxFeeFromBlockCalled                  func(tx data.TransactionHandler)
	GetAccumulatedFeesCalled                 func() *big.Int
	GetDeveloperFeesCalled                   func() *big.Int
	RevertFeesCalled                         func(txHashes [][]byte)
}

// RevertFees -
func (ut *UnsignedTxHandlerStub) RevertFees(txHashes [][]byte) {
	if ut.RevertFeesCalled != nil {
		ut.RevertFeesCalled(txHashes)
	}
}

// CreateBlockStarted -
func (ut *UnsignedTxHandlerStub) CreateBlockStarted(_ scheduled.GasAndFees) {
}

// GetAccumulatedFees -
func (ut *UnsignedTxHandlerStub) GetAccumulatedFees() *big.Int {
	if ut.GetAccumulatedFeesCalled != nil {
		return ut.GetAccumulatedFeesCalled()
	}
	return big.NewInt(0)
}

// GetDeveloperFees -
func (ut *UnsignedTxHandlerStub) GetDeveloperFees() *big.Int {
	if ut.GetDeveloperFeesCalled != nil {
		return ut.GetDeveloperFeesCalled()
	}
	return big.NewInt(0)
}

// AddRewardTxFromBlock -
func (ut *UnsignedTxHandlerStub) AddRewardTxFromBlock(tx data.TransactionHandler) {
	if ut.AddTxFeeFromBlockCalled == nil {
		return
	}

	ut.AddTxFeeFromBlockCalled(tx)
}

// CleanProcessedUTxs -
func (ut *UnsignedTxHandlerStub) CleanProcessedUTxs() {
	if ut.CleanProcessedUtxsCalled == nil {
		return
	}

	ut.CleanProcessedUtxsCalled()
}

// ProcessTransactionFee -
func (ut *UnsignedTxHandlerStub) ProcessTransactionFee(cost *big.Int, devFee *big.Int, txHash []byte) {
	if ut.ProcessTransactionFeeCalled == nil {
		return
	}

	ut.ProcessTransactionFeeCalled(cost, devFee, txHash)
}

// ProcessTransactionFeeRelayedUserTx -
func (ut *UnsignedTxHandlerStub) ProcessTransactionFeeRelayedUserTx(cost *big.Int, devFee *big.Int, userTxHash []byte, originalTxHash []byte) {
	if ut.ProcessTransactionFeeRelayedUserTxCalled == nil {
		return
	}

	ut.ProcessTransactionFeeRelayedUserTxCalled(cost, devFee, userTxHash, originalTxHash)
}

// CreateAllUTxs -
func (ut *UnsignedTxHandlerStub) CreateAllUTxs() []data.TransactionHandler {
	if ut.CreateAllUTxsCalled == nil {
		return nil
	}
	return ut.CreateAllUTxsCalled()
}

// VerifyCreatedUTxs -
func (ut *UnsignedTxHandlerStub) VerifyCreatedUTxs() error {
	if ut.VerifyCreatedUTxsCalled == nil {
		return nil
	}
	return ut.VerifyCreatedUTxsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ut *UnsignedTxHandlerStub) IsInterfaceNil() bool {
	return ut == nil
}
