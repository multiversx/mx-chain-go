package sovereign

import "github.com/multiversx/mx-chain-core-go/data"

// OutgoingOperationsFormatterMock -
type OutgoingOperationsFormatterMock struct {
	CreateOutgoingTxDataCalled func(logs []*data.LogData) [][]byte
}

// CreateOutgoingTxsData -
func (stub *OutgoingOperationsFormatterMock) CreateOutgoingTxsData(logs []*data.LogData) [][]byte {
	if stub.CreateOutgoingTxDataCalled != nil {
		return stub.CreateOutgoingTxDataCalled(logs)
	}

	return nil
}

// IsInterfaceNil -
func (stub *OutgoingOperationsFormatterMock) IsInterfaceNil() bool {
	return stub == nil
}
