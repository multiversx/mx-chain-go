package mock

import (
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/state"
	sovereignMocks "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	AccountCreator         state.AccountFactory
	OutGoingOperationsPool block.OutGoingOperationsPool
	DataCodec              sovereign.DataDecoderHandler
	TopicsChecker          sovereign.TopicsCheckerHandler
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		AccountCreator:         &stateMock.AccountsFactoryStub{},
		OutGoingOperationsPool: &sovereignMocks.OutGoingOperationsPoolMock{},
		DataCodec:              &sovereignMocks.DataCodecMock{},
		TopicsChecker:          &sovereignMocks.TopicsCheckerMock{},
	}
}

// AccountsCreator  -
func (r *RunTypeComponentsStub) AccountsCreator() state.AccountFactory {
	return r.AccountCreator
}

func (r *RunTypeComponentsStub) OutGoingOperationsPoolHandler() block.OutGoingOperationsPool {
	return r.OutGoingOperationsPool
}

// DataCodecHandler  -
func (r *RunTypeComponentsStub) DataCodecHandler() sovereign.DataDecoderHandler {
	return r.DataCodec
}

// TopicsCheckerHandler  -
func (r *RunTypeComponentsStub) TopicsCheckerHandler() sovereign.TopicsCheckerHandler {
	return r.TopicsChecker
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
