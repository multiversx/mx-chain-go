package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// PostProcessorTxsStub -
type PostProcessorTxsStub struct {
	InitCalled                                 func()
	AddPostProcessorTxCalled                   func([]byte) bool
	IsPostProcessorTxAddedCalled               func([]byte) bool
	SetTransactionCoordinatorCalled            func(process.TransactionCoordinator)
	GetAllIntermediateTxsHashesForTxHashCalled func([]byte) map[block.Type]map[uint32][][]byte
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

// GetAllIntermediateTxsHashesForTxHash -
func (ppts *PostProcessorTxsStub) GetAllIntermediateTxsHashesForTxHash(txHash []byte) map[block.Type]map[uint32][][]byte {
	if ppts.GetAllIntermediateTxsHashesForTxHashCalled != nil {
		return ppts.GetAllIntermediateTxsHashesForTxHashCalled(txHash)
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
