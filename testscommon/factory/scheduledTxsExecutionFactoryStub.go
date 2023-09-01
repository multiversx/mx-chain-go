package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

// ScheduledTxsExecutionFactoryStub -
type ScheduledTxsExecutionFactoryStub struct {
	CreateScheduledTxsExecutionHandlerCalled func(args preprocess.ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error)
}

// CreateScheduledTxsExecutionHandler -
func (s *ScheduledTxsExecutionFactoryStub) CreateScheduledTxsExecutionHandler(args preprocess.ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error) {
	if s.CreateScheduledTxsExecutionHandlerCalled != nil {
		return s.CreateScheduledTxsExecutionHandlerCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (s *ScheduledTxsExecutionFactoryStub) IsInterfaceNil() bool {
	return s == nil
}
