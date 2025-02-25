package sovereign

import "github.com/multiversx/mx-chain-core-go/data"

// OutgoingOperationsFormatterMock -
type OutgoingOperationsFormatterMock struct {
	CreateOutgoingTxDataCalled              func(logs []*data.LogData) ([][]byte, error)
	CreateOutGoingChangeValidatorDataCalled func(pubKeys []string) ([]byte, error)
}

// CreateOutgoingTxsData -
func (stub *OutgoingOperationsFormatterMock) CreateOutgoingTxsData(logs []*data.LogData) ([][]byte, error) {
	if stub.CreateOutgoingTxDataCalled != nil {
		return stub.CreateOutgoingTxDataCalled(logs)
	}

	return make([][]byte, 0), nil
}

// CreateOutGoingChangeValidatorData -
func (stub *OutgoingOperationsFormatterMock) CreateOutGoingChangeValidatorData(pubKeys []string) ([]byte, error) {
	if stub.CreateOutGoingChangeValidatorDataCalled != nil {
		return stub.CreateOutGoingChangeValidatorDataCalled(pubKeys)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (stub *OutgoingOperationsFormatterMock) IsInterfaceNil() bool {
	return stub == nil
}
