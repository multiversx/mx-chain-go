package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// ScheduledTxsExecutionFactoryMock -
type ScheduledTxsExecutionFactoryMock struct {
	CreateScheduledTxsExecutionHandlerCalled func(args preprocess.ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error)
}

// CreateScheduledTxsExecutionHandler -
func (s *ScheduledTxsExecutionFactoryMock) CreateScheduledTxsExecutionHandler(args preprocess.ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error) {
	if s.CreateScheduledTxsExecutionHandlerCalled != nil {
		return s.CreateScheduledTxsExecutionHandlerCalled(args)
	}
	return &testscommon.ScheduledTxsExecutionStub{}, nil
}

// IsInterfaceNil -
func (s *ScheduledTxsExecutionFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
