package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/estimator"
)

// InclusionEstimatorMock -
type InclusionEstimatorMock struct {
	DecideCalled func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.ExecutionResultHandler, currentHdrTsMs uint64) (allowed int)
}

// Decide -
func (mock *InclusionEstimatorMock) Decide(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.ExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
	if mock.DecideCalled != nil {
		return mock.DecideCalled(lastNotarised, pending, currentHdrTsMs)
	}

	return 0
}

// IsInterfaceNil -
func (mock *InclusionEstimatorMock) IsInterfaceNil() bool {
	return mock == nil
}
