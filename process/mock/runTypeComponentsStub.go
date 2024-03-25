package mock

import (
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/state"
	sovereignMocks "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	AccountCreator       state.AccountFactory
	DataCodecFactory     sovereign.DataDecoderCreator
	TopicsCheckerFactory sovereign.TopicsCheckerCreator
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		AccountCreator:       &stateMock.AccountsFactoryStub{},
		DataCodecFactory:     &sovereignMocks.DataCodecFactoryMock{},
		TopicsCheckerFactory: &sovereignMocks.TopicsCheckerFactoryMock{},
	}
}

// AccountsCreator  -
func (r *RunTypeComponentsStub) AccountsCreator() state.AccountFactory {
	return r.AccountCreator
}

// DataCodecCreator  -
func (r *RunTypeComponentsStub) DataCodecCreator() sovereign.DataDecoderCreator {
	return r.DataCodecFactory
}

// TopicsCheckerCreator  -
func (r *RunTypeComponentsStub) TopicsCheckerCreator() sovereign.TopicsCheckerCreator {
	return r.TopicsCheckerFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
