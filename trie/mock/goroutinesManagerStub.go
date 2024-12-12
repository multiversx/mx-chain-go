package mock

import "github.com/multiversx/mx-chain-go/common"

// GoroutinesManagerStub -
type GoroutinesManagerStub struct {
	ShouldContinueProcessingCalled func() bool
	CanStartGoRoutineCalled        func() bool
	EndGoRoutineProcessingCalled   func()
	SetNewErrorChannelCalled       func(common.BufferedErrChan) error
	SetErrorCalled                 func(error)
	GetErrorCalled                 func() error
}

// ShouldContinueProcessing -
func (g *GoroutinesManagerStub) ShouldContinueProcessing() bool {
	if g.ShouldContinueProcessingCalled != nil {
		return g.ShouldContinueProcessingCalled()
	}
	return true
}

// CanStartGoRoutine -
func (g *GoroutinesManagerStub) CanStartGoRoutine() bool {
	if g.CanStartGoRoutineCalled != nil {
		return g.CanStartGoRoutineCalled()
	}
	return true
}

// EndGoRoutineProcessing -
func (g *GoroutinesManagerStub) EndGoRoutineProcessing() {
	if g.EndGoRoutineProcessingCalled != nil {
		g.EndGoRoutineProcessingCalled()
	}
}

// SetNewErrorChannel -
func (g *GoroutinesManagerStub) SetNewErrorChannel(errChan common.BufferedErrChan) error {
	if g.SetNewErrorChannelCalled != nil {
		return g.SetNewErrorChannelCalled(errChan)
	}
	return nil
}

// SetError -
func (g *GoroutinesManagerStub) SetError(err error) {
	if g.SetErrorCalled != nil {
		g.SetErrorCalled(err)
	}
}

// GetError -
func (g *GoroutinesManagerStub) GetError() error {
	if g.GetErrorCalled != nil {
		return g.GetErrorCalled()
	}
	return nil
}

// IsInterfaceNil -
func (g *GoroutinesManagerStub) IsInterfaceNil() bool {
	return g == nil
}
