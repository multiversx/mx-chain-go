package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ScheduledTxsExecutionStub -
type ScheduledTxsExecutionStub struct {
	InitCalled                    func()
	AddCalled                     func([]byte, data.TransactionHandler) bool
	ExecuteCalled                 func([]byte) error
	ExecuteAllCalled              func(func() time.Duration, process.TransactionCoordinator) error
	GetScheduledSCRsCalled        func() map[block.Type][]data.TransactionHandler
	SetScheduledSCRsCalled        func(map[block.Type][]data.TransactionHandler)
	SetTransactionProcessorCalled func(txProcessor process.TransactionProcessor)
}

// Init -
func (stes *ScheduledTxsExecutionStub) Init() {
	if stes.InitCalled != nil {
		stes.InitCalled()
	}
}

// Add -
func (stes *ScheduledTxsExecutionStub) Add(txHash []byte, tx data.TransactionHandler) bool {
	if stes.AddCalled != nil {
		return stes.AddCalled(txHash, tx)
	}
	return true
}

// Execute -
func (stes *ScheduledTxsExecutionStub) Execute(txHash []byte) error {
	if stes.ExecuteCalled != nil {
		return stes.ExecuteCalled(txHash)
	}
	return nil
}

// ExecuteAll -
func (stes *ScheduledTxsExecutionStub) ExecuteAll(haveTime func() time.Duration, txCoordinator process.TransactionCoordinator) error {
	if stes.ExecuteAllCalled != nil {
		return stes.ExecuteAllCalled(haveTime, txCoordinator)
	}
	return nil
}

// GetScheduledSCRs -
func (stes *ScheduledTxsExecutionStub) GetScheduledSCRs() map[block.Type][]data.TransactionHandler {
	if stes.GetScheduledSCRsCalled != nil {
		return stes.GetScheduledSCRsCalled()
	}
	return nil
}

// SetScheduledSCRs -
func (stes *ScheduledTxsExecutionStub) SetScheduledSCRs(mapScheduledSCRs map[block.Type][]data.TransactionHandler) {
	if stes.SetScheduledSCRsCalled != nil {
		stes.SetScheduledSCRsCalled(mapScheduledSCRs)
	}
}

// SetTransactionProcessor -
func (stes *ScheduledTxsExecutionStub) SetTransactionProcessor(txProcessor process.TransactionProcessor) {
	if stes.SetTransactionProcessorCalled != nil {
		stes.SetTransactionProcessorCalled(txProcessor)
	}
}

// IsInterfaceNil -
func (stes *ScheduledTxsExecutionStub) IsInterfaceNil() bool {
	return stes == nil
}
