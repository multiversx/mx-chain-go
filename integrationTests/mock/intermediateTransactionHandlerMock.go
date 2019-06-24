package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type IntermediateTransactionHandlerMock struct {
	AddIntermediateTransactionsCalled func(txs []data.TransactionHandler) error
	CreateAllInterMiniBlocksCalled    func() []*block.MiniBlock
	VerifyInterMiniBlocksCalled       func(body block.Body) error
}

func (ith *IntermediateTransactionHandlerMock) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	if ith.AddIntermediateTransactionsCalled == nil {
		return nil
	}
	return ith.AddIntermediateTransactionsCalled(txs)
}

func (ith *IntermediateTransactionHandlerMock) CreateAllInterMiniBlocks() []*block.MiniBlock {
	if ith.CreateAllInterMiniBlocksCalled == nil {
		return nil
	}
	return ith.CreateAllInterMiniBlocksCalled()
}

func (ith *IntermediateTransactionHandlerMock) VerifyInterMiniBlocks(body block.Body) error {
	if ith.VerifyInterMiniBlocksCalled == nil {
		return nil
	}
	return ith.VerifyInterMiniBlocksCalled(body)
}
