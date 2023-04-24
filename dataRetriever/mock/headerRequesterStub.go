package mock

import "github.com/multiversx/mx-chain-go/dataRetriever"

// HeaderRequesterStub -
type HeaderRequesterStub struct {
	RequestDataFromHashCalled func(hash []byte, epoch uint32) error
	SetNumPeersToQueryCalled  func(intra int, cross int)
	NumPeersToQueryCalled     func() (int, int)
	SetDebugHandlerCalled     func(handler dataRetriever.DebugHandler) error
	SetEpochHandlerCalled     func(epochHandler dataRetriever.EpochHandler) error
}

// RequestDataFromHash -
func (stub *HeaderRequesterStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if stub.RequestDataFromHashCalled != nil {
		return stub.RequestDataFromHashCalled(hash, epoch)
	}

	return nil
}

// SetNumPeersToQuery -
func (stub *HeaderRequesterStub) SetNumPeersToQuery(intra int, cross int) {
	if stub.SetNumPeersToQueryCalled != nil {
		stub.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (stub *HeaderRequesterStub) NumPeersToQuery() (int, int) {
	if stub.NumPeersToQueryCalled != nil {
		return stub.NumPeersToQueryCalled()
	}

	return 0, 0
}

// SetDebugHandler -
func (stub *HeaderRequesterStub) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	if stub.SetDebugHandlerCalled != nil {
		return stub.SetDebugHandlerCalled(handler)
	}

	return nil
}

// SetEpochHandler -
func (stub *HeaderRequesterStub) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if stub.SetEpochHandlerCalled != nil {
		return stub.SetEpochHandlerCalled(epochHandler)
	}

	return nil
}

// IsInterfaceNil -
func (stub *HeaderRequesterStub) IsInterfaceNil() bool {
	return stub == nil
}
