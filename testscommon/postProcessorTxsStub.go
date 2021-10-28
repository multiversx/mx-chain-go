package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// PostProcessorTxsStub -
type PostProcessorTxsStub struct {
	InitCalled                      func()
	AddPostProcessorTxCalled        func([]byte, data.TransactionHandler) bool
	GetPostProcessorTxCalled        func([]byte) data.TransactionHandler
	IsPostProcessorTxAddedCalled    func([]byte) bool
	SetTransactionCoordinatorCalled func(process.TransactionCoordinator)
	GetProcessedResultsCalled       func() map[block.Type]map[uint32][]*process.TxInfo
	InitProcessedResultsCalled      func()
}

// Init -
func (ppts *PostProcessorTxsStub) Init() {
	if ppts.InitCalled != nil {
		ppts.InitCalled()
	}
}

// AddPostProcessorTx -
func (ppts *PostProcessorTxsStub) AddPostProcessorTx(txHash []byte, txHandler data.TransactionHandler) bool {
	if ppts.AddPostProcessorTxCalled != nil {
		return ppts.AddPostProcessorTxCalled(txHash, txHandler)
	}
	return true
}

// GetPostProcessorTx -
func (ppts *PostProcessorTxsStub) GetPostProcessorTx(txHash []byte) data.TransactionHandler {
	if ppts.GetPostProcessorTxCalled != nil {
		return ppts.GetPostProcessorTxCalled(txHash)
	}
	return nil
}

// SetTransactionCoordinator -
func (ppts *PostProcessorTxsStub) SetTransactionCoordinator(txCoordinator process.TransactionCoordinator) {
	if ppts.SetTransactionCoordinatorCalled != nil {
		ppts.SetTransactionCoordinatorCalled(txCoordinator)
	}
}

// GetProcessedResults -
func (ppts *PostProcessorTxsStub) GetProcessedResults() map[block.Type]map[uint32][]*process.TxInfo {
	if ppts.GetProcessedResultsCalled != nil {
		return ppts.GetProcessedResultsCalled()
	}
	return nil
}

// InitProcessedResults -
func (ppts *PostProcessorTxsStub) InitProcessedResults() {
	if ppts.InitProcessedResultsCalled != nil {
		ppts.InitProcessedResultsCalled()
	}
}

// IsPostProcessorTxAdded -
func (ppts *PostProcessorTxsStub) IsPostProcessorTxAdded(txHash []byte) bool {
	if ppts.IsPostProcessorTxAddedCalled != nil {
		return ppts.IsPostProcessorTxAddedCalled(txHash)
	}
	return false
}

// IsInterfaceNil -
func (ppts *PostProcessorTxsStub) IsInterfaceNil() bool {
	return ppts == nil
}
