package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type IntermediateTransactionHandlerMock struct {
	AddIntermediateTransactionsCalled        func(txs []data.TransactionHandler) error
	CreateAllInterMiniBlocksCalled           func() map[uint32]*block.MiniBlock
	VerifyInterMiniBlocksCalled              func(body block.Body) error
	SaveCurrentIntermediateTxToStorageCalled func() error
	CreateBlockStartedCalled                 func()
	CreateMarshalizedDataCalled              func(txHashes [][]byte) ([][]byte, error)
	GetAllCurrentFinishedTxsCalled           func() map[string]data.TransactionHandler
}

func (ith *IntermediateTransactionHandlerMock) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	if ith.CreateMarshalizedDataCalled == nil {
		return nil, nil
	}
	return ith.CreateMarshalizedDataCalled(txHashes)
}

func (ith *IntermediateTransactionHandlerMock) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	if ith.AddIntermediateTransactionsCalled == nil {
		return nil
	}
	return ith.AddIntermediateTransactionsCalled(txs)
}

func (ith *IntermediateTransactionHandlerMock) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
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

func (ith *IntermediateTransactionHandlerMock) SaveCurrentIntermediateTxToStorage() error {
	if ith.SaveCurrentIntermediateTxToStorageCalled == nil {
		return nil
	}
	return ith.SaveCurrentIntermediateTxToStorageCalled()
}

func (ith *IntermediateTransactionHandlerMock) CreateBlockStarted() {
	if ith.CreateBlockStartedCalled != nil {
		ith.CreateAllInterMiniBlocksCalled()
	}
}

func (ith *IntermediateTransactionHandlerMock) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
	if ith.GetAllCurrentFinishedTxsCalled != nil {
		return ith.GetAllCurrentFinishedTxsCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ith *IntermediateTransactionHandlerMock) IsInterfaceNil() bool {
	if ith == nil {
		return true
	}
	return false
}
