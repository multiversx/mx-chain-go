package testscommon

import "github.com/multiversx/mx-chain-core-go/data/endProcess"

// ShuffleOutCloserStub -
type ShuffleOutCloserStub struct {
	EndOfProcessingHandlerCalled func(event endProcess.ArgEndProcess) error
	CloseCalled                  func() error
}

// EndOfProcessingHandler -
func (stub *ShuffleOutCloserStub) EndOfProcessingHandler(event endProcess.ArgEndProcess) error {
	if stub.EndOfProcessingHandlerCalled != nil {
		return stub.EndOfProcessingHandlerCalled(event)
	}
	return nil
}

// IsInterfaceNil -
func (stub *ShuffleOutCloserStub) IsInterfaceNil() bool {
	return stub == nil
}

// Close -
func (stub *ShuffleOutCloserStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}
