package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
)

// InclusionEstimatorMock -
type InclusionEstimatorMock struct {
	DecideCalled func(lastNotarised *common.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int)
}

// Decide -
func (mock *InclusionEstimatorMock) Decide(lastNotarised *common.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
	if mock.DecideCalled != nil {
		return mock.DecideCalled(lastNotarised, pending, currentHdrTsMs)
	}

	return 0
}

// IsInterfaceNil -
func (mock *InclusionEstimatorMock) IsInterfaceNil() bool {
	return mock == nil
}
