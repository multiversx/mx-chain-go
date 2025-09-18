package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

// ExecutionResultsVerifierMock -
type ExecutionResultsVerifierMock struct {
	VerifyHeaderExecutionResultsCalled func(headerHash []byte, header data.HeaderHandler) error
}

// VerifyHeaderExecutionResults -
func (mock *ExecutionResultsVerifierMock) VerifyHeaderExecutionResults(headerHash []byte, header data.HeaderHandler) error {
	if mock.VerifyHeaderExecutionResultsCalled != nil {
		return mock.VerifyHeaderExecutionResultsCalled(headerHash, header)
	}
	return nil
}

// IsInterfaceNil -
func (mock *ExecutionResultsVerifierMock) IsInterfaceNil() bool {
	return mock == nil
}
