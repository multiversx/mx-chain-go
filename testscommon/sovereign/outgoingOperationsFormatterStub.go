package sovereign

import "github.com/multiversx/mx-chain-core-go/data"

// OutgoingOperationsFormatterStub -
type OutgoingOperationsFormatterStub struct {
	CreateOutgoingTxDataCalled func(logs []*data.LogData) []byte
}

// CreateOutgoingTxData -
func (stub *OutgoingOperationsFormatterStub) CreateOutgoingTxData(logs []*data.LogData) []byte {
	if stub.CreateOutgoingTxDataCalled != nil {
		return stub.CreateOutgoingTxDataCalled(logs)
	}

	return nil
}

// IsInterfaceNil -
func (stub *OutgoingOperationsFormatterStub) IsInterfaceNil() bool {
	return stub == nil
}
