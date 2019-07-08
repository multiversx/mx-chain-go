package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type UnsignedTxHandlerMock struct {
	CleanProcessedUtxsCalled func()
	AddProcessedUTxCalled    func(tx data.TransactionHandler)
	CreateAllUTxsCalled      func() []data.TransactionHandler
	VerifyCreatedUTxsCalled  func(header data.HeaderHandler, body block.Body) error
}

func (ut *UnsignedTxHandlerMock) CleanProcessedUTxs() {
	if ut.CleanProcessedUtxsCalled == nil {
		return
	}

	ut.CleanProcessedUtxsCalled()
	return
}

func (ut *UnsignedTxHandlerMock) AddProcessedUTx(tx data.TransactionHandler) {
	if ut.AddProcessedUTxCalled == nil {
		return
	}

	ut.AddProcessedUTxCalled(tx)
	return
}

func (ut *UnsignedTxHandlerMock) CreateAllUTxs() []data.TransactionHandler {
	if ut.CreateAllUTxsCalled == nil {
		return nil
	}
	return ut.CreateAllUTxsCalled()
}

func (ut *UnsignedTxHandlerMock) VerifyCreatedUTxs(header data.HeaderHandler, body block.Body) error {
	if ut.VerifyCreatedUTxsCalled == nil {
		return nil
	}
	return ut.VerifyCreatedUTxsCalled(header, body)
}
