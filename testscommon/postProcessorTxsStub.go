package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// PostProcessorTxsStub -
type PostProcessorTxsStub struct {
	InitCalled                           func()
	AddPostProcessorTxCalled             func([]byte) bool
	IsPostProcessorTxAddedCalled         func([]byte) bool
	SetTransactionCoordinatorCalled      func(process.TransactionCoordinator)
	GetAllIntermediateTxsForTxHashCalled func([]byte) map[block.Type]map[uint32][]*process.TxInfo
}

// Init -
func (ppts *PostProcessorTxsStub) Init() {
	if ppts.InitCalled != nil {
		ppts.InitCalled()
	}
}

// AddPostProcessorTx -
func (ppts *PostProcessorTxsStub) AddPostProcessorTx(txHash []byte) bool {
	if ppts.AddPostProcessorTxCalled != nil {
		return ppts.AddPostProcessorTxCalled(txHash)
	}
	return true
}

// SetTransactionCoordinator -
func (ppts *PostProcessorTxsStub) SetTransactionCoordinator(txCoordinator process.TransactionCoordinator) {
	if ppts.SetTransactionCoordinatorCalled != nil {
		ppts.SetTransactionCoordinatorCalled(txCoordinator)
	}
}

// GetAllIntermediateTxsForTxHash -
func (ppts *PostProcessorTxsStub) GetAllIntermediateTxsForTxHash(txHash []byte) map[block.Type]map[uint32][]*process.TxInfo {
	if ppts.GetAllIntermediateTxsForTxHashCalled != nil {
		return ppts.GetAllIntermediateTxsForTxHashCalled(txHash)
	}
	return nil
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
