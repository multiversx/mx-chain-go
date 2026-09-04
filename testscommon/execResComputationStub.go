package testscommon

import "github.com/multiversx/mx-chain-go/process/estimator"

// ExecResSizeLimitCheckerStub -
type ExecResSizeLimitCheckerStub struct {
	AddNumExecResCalled           func(numExecRes int)
	IsMaxExecResSizeReachedCalled func(numNewExecRes int) bool
}

// AddNumExecRes -
func (stub *ExecResSizeLimitCheckerStub) AddNumExecRes(numExecRes int) {
	if stub.AddNumExecResCalled != nil {
		stub.AddNumExecResCalled(numExecRes)
	}
}

// IsMaxExecResSizeReached -
func (stub *ExecResSizeLimitCheckerStub) IsMaxExecResSizeReached(numNewExecRes int) bool {
	if stub.IsMaxExecResSizeReachedCalled != nil {
		return stub.IsMaxExecResSizeReachedCalled(numNewExecRes)
	}

	return false
}

// IsInterfaceNil -
func (stub *ExecResSizeLimitCheckerStub) IsInterfaceNil() bool {
	return stub == nil
}

// ExecResSizeComputationStub -
type ExecResSizeComputationStub struct {
	NewComputationCalled func() estimator.ExecResSizeLimitCheckerHandler
}

// NewComputation -
func (stub *ExecResSizeComputationStub) NewComputation() estimator.ExecResSizeLimitCheckerHandler {
	return &ExecResSizeLimitCheckerStub{}
}

// IsInterfaceNil -
func (stub *ExecResSizeComputationStub) IsInterfaceNil() bool {
	return stub == nil
}
