package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// IntermediateTransactionHandlerStub -
type IntermediateTransactionHandlerStub struct {
	AddIntermediateTransactionsCalled        func(txs []data.TransactionHandler) error
	GetNumOfCrossInterMbsAndTxsCalled        func() (int, int)
	CreateAllInterMiniBlocksCalled           func() []*block.MiniBlock
	VerifyInterMiniBlocksCalled              func(body *block.Body) error
	SaveCurrentIntermediateTxToStorageCalled func()
	CreateBlockStartedCalled                 func()
	CreateMarshalledDataCalled               func(txHashes [][]byte) ([][]byte, error)
	GetAllCurrentFinishedTxsCalled           func() map[string]data.TransactionHandler
	RemoveProcessedResultsCalled             func(key []byte) [][]byte
	InitProcessedResultsCalled               func(key []byte)
	intermediateTransactions                 []data.TransactionHandler
}

// RemoveProcessedResults -
func (ith *IntermediateTransactionHandlerStub) RemoveProcessedResults(key []byte) [][]byte {
	if ith.RemoveProcessedResultsCalled != nil {
		return ith.RemoveProcessedResultsCalled(key)
	}
	return nil
}

// InitProcessedResults -
func (ith *IntermediateTransactionHandlerStub) InitProcessedResults(key []byte) {
	if ith.InitProcessedResultsCalled != nil {
		ith.InitProcessedResultsCalled(key)
	}
}

// CreateMarshalledData -
func (ith *IntermediateTransactionHandlerStub) CreateMarshalledData(txHashes [][]byte) ([][]byte, error) {
	if ith.CreateMarshalledDataCalled == nil {
		return nil, nil
	}
	return ith.CreateMarshalledDataCalled(txHashes)
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

// GetNumOfCrossInterMbsAndTxs -
func (ith *IntermediateTransactionHandlerStub) GetNumOfCrossInterMbsAndTxs() (int, int) {
	if ith.GetNumOfCrossInterMbsAndTxsCalled == nil {
		return 0, 0
	}
	return ith.GetNumOfCrossInterMbsAndTxsCalled()
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
func (ith *IntermediateTransactionHandlerStub) SaveCurrentIntermediateTxToStorage() {
	if ith.SaveCurrentIntermediateTxToStorageCalled == nil {
		return
	}
	ith.SaveCurrentIntermediateTxToStorageCalled()
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
