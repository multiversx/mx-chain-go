package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

type UnsignedTxHandlerMock struct {
	CleanProcessedUtxsCalled    func()
	ProcessTransactionFeeCalled func(cost *big.Int)
	CreateAllUTxsCalled         func() []data.TransactionHandler
	VerifyCreatedUTxsCalled     func() error
	AddTxFeeFromBlockCalled     func(tx data.TransactionHandler)
}

func (ut *UnsignedTxHandlerMock) AddRewardTxFromBlock(tx data.TransactionHandler) {
	if ut.AddTxFeeFromBlockCalled == nil {
		return
	}

	ut.AddTxFeeFromBlockCalled(tx)
}

func (ut *UnsignedTxHandlerMock) CleanProcessedUTxs() {
	if ut.CleanProcessedUtxsCalled == nil {
		return
	}

	ut.CleanProcessedUtxsCalled()
}

func (ut *UnsignedTxHandlerMock) ProcessTransactionFee(cost *big.Int) {
	if ut.ProcessTransactionFeeCalled == nil {
		return
	}

	ut.ProcessTransactionFeeCalled(cost)
}

func (ut *UnsignedTxHandlerMock) CreateAllUTxs() []data.TransactionHandler {
	if ut.CreateAllUTxsCalled == nil {
		return nil
	}
	return ut.CreateAllUTxsCalled()
}

func (ut *UnsignedTxHandlerMock) VerifyCreatedUTxs() error {
	if ut.VerifyCreatedUTxsCalled == nil {
		return nil
	}
	return ut.VerifyCreatedUTxsCalled()
}
