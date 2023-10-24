package sovereign

import "github.com/multiversx/mx-chain-core-go/data"

type OutgoingOperationsFormatterStub struct {
	CreateOutgoingTxDataCalled func(logs []*data.LogData) []byte
}

func (stub *OutgoingOperationsFormatterStub) CreateOutgoingTxData(logs []*data.LogData) []byte {
	if stub.CreateOutgoingTxDataCalled != nil {
		return stub.CreateOutgoingTxDataCalled(logs)
	}

	return nil
}

func (stub *OutgoingOperationsFormatterStub) IsInterfaceNil() bool {
	return stub == nil
}
