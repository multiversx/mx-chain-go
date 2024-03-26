package mock

import (
	factorySovereign "github.com/multiversx/mx-chain-go/factory/sovereign"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

// RunTypeComponentsStub -
type RunTypeComponentsStub struct {
	AccountCreator       state.AccountFactory
	DataCodecFactory     factorySovereign.DataDecoderCreator
	TopicsCheckerFactory factorySovereign.TopicsCheckerCreator
}

// NewRunTypeComponentsStub -
func NewRunTypeComponentsStub() *RunTypeComponentsStub {
	return &RunTypeComponentsStub{
		AccountCreator:       &stateMock.AccountsFactoryStub{},
		DataCodecFactory:     &genericMocks.DataCodecFactoryMock{},
		TopicsCheckerFactory: &genericMocks.TopicsCheckerFactoryMock{},
	}
}

// AccountsCreator  -
func (r *RunTypeComponentsStub) AccountsCreator() state.AccountFactory {
	return r.AccountCreator
}

// DataCodecCreator  -
func (r *RunTypeComponentsStub) DataCodecCreator() factorySovereign.DataDecoderCreator {
	return r.DataCodecFactory
}

// TopicsCheckerCreator  -
func (r *RunTypeComponentsStub) TopicsCheckerCreator() factorySovereign.TopicsCheckerCreator {
	return r.TopicsCheckerFactory
}

// IsInterfaceNil -
func (r *RunTypeComponentsStub) IsInterfaceNil() bool {
	return r == nil
}
