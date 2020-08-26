package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// IntermediateTransactionHandlerStub -
type IntermediateTransactionHandlerStub struct {
	AddIntermediateTransactionsCalled        func(txs []data.TransactionHandler) error
	CreateAllInterMiniBlocksCalled           func() []*block.MiniBlock
	VerifyInterMiniBlocksCalled              func(body *block.Body) error
	SaveCurrentIntermediateTxToStorageCalled func() error
	CreateBlockStartedCalled                 func()
	CreateMarshalizedDataCalled              func(txHashes [][]byte) ([][]byte, error)
	GetAllCurrentFinishedTxsCalled           func() map[string]data.TransactionHandler
	RemoveProcessedResultsForCalled          func(txHashes [][]byte)
	intermediateTransactions                 []data.TransactionHandler
}

// RemoveProcessedResultsFor -
func (ith *IntermediateTransactionHandlerStub) RemoveProcessedResultsFor(txHashes [][]byte) {
	if ith.RemoveProcessedResultsForCalled != nil {
		ith.RemoveProcessedResultsForCalled(txHashes)
	}
}

// CreateMarshalizedData -
func (ith *IntermediateTransactionHandlerStub) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	if ith.CreateMarshalizedDataCalled == nil {
		return nil, nil
	}
	return ith.CreateMarshalizedDataCalled(txHashes)
}

// AddIntermediateTransactions -
func (ith *IntermediateTransactionHandlerStub) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	if ith.AddIntermediateTransactionsCalled == nil {
		ith.intermediateTransactions = append(ith.intermediateTransactions, txs...)
		return nil
	}
	return ith.AddIntermediateTransactionsCalled(txs)
}

// GetIntermediateTransactions -
func (ith *IntermediateTransactionHandlerStub) GetIntermediateTransactions() []data.TransactionHandler {
	return ith.intermediateTransactions
}

// CreateAllInterMiniBlocks -
func (ith *IntermediateTransactionHandlerStub) CreateAllInterMiniBlocks() []*block.MiniBlock {
	if ith.CreateAllInterMiniBlocksCalled == nil {
		return nil
	}
	return ith.CreateAllInterMiniBlocksCalled()
}

// VerifyInterMiniBlocks -
func (ith *IntermediateTransactionHandlerStub) VerifyInterMiniBlocks(body *block.Body) error {
	if ith.VerifyInterMiniBlocksCalled == nil {
		return nil
	}
	return ith.VerifyInterMiniBlocksCalled(body)
}

// SaveCurrentIntermediateTxToStorage -
func (ith *IntermediateTransactionHandlerStub) SaveCurrentIntermediateTxToStorage() error {
	if ith.SaveCurrentIntermediateTxToStorageCalled == nil {
		return nil
	}
	return ith.SaveCurrentIntermediateTxToStorageCalled()
}

// CreateBlockStarted -
func (ith *IntermediateTransactionHandlerStub) CreateBlockStarted() {
	if ith.CreateBlockStartedCalled != nil {
		ith.CreateBlockStartedCalled()
	}
}

// GetAllCurrentFinishedTxs -
func (ith *IntermediateTransactionHandlerStub) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
	if ith.GetAllCurrentFinishedTxsCalled != nil {
		return ith.GetAllCurrentFinishedTxsCalled()
	}
	return nil
}

// GetCreatedInShardMiniBlock -
func (ith *IntermediateTransactionHandlerStub) GetCreatedInShardMiniBlock() *block.MiniBlock {
	return &block.MiniBlock{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ith *IntermediateTransactionHandlerStub) IsInterfaceNil() bool {
	return ith == nil
}
