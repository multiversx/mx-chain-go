package dataRetriever

import "github.com/multiversx/mx-chain-go/dataRetriever"

// HeaderRequesterStub -
type HeaderRequesterStub struct {
	RequestDataFromHashCalled  func(hash []byte, epoch uint32) error
	SetNumPeersToQueryCalled   func(intra int, cross int)
	NumPeersToQueryCalled      func() (int, int)
	SetDebugHandlerCalled      func(handler dataRetriever.DebugHandler) error
	RequestDataFromEpochCalled func(identifier []byte) error
	RequestDataFromNonceCalled func(nonce uint64, epoch uint32) error
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

// RequestDataFromEpoch -
func (stub *HeaderRequesterStub) RequestDataFromEpoch(identifier []byte) error {
	if stub.RequestDataFromEpochCalled != nil {
		return stub.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

// RequestDataFromNonce -
func (stub *HeaderRequesterStub) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	if stub.RequestDataFromNonceCalled != nil {
		return stub.RequestDataFromNonceCalled(nonce, epoch)
	}
	return nil
}

// IsInterfaceNil -
func (stub *HeaderRequesterStub) IsInterfaceNil() bool {
	return stub == nil
}
