package dataRetriever

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

// RequesterStub -
type RequesterStub struct {
	RequestDataFromHashCalled     func(hash []byte, epoch uint32) error
	SetNumPeersToQueryCalled      func(intra int, cross int)
	NumPeersToQueryCalled         func() (int, int)
	SetResolverDebugHandlerCalled func(handler dataRetriever.ResolverDebugHandler) error
}

// RequestDataFromHash -
func (stub *RequesterStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if stub.RequestDataFromHashCalled != nil {
		return stub.RequestDataFromHashCalled(hash, epoch)
	}
	return nil
}

// SetNumPeersToQuery -
func (stub *RequesterStub) SetNumPeersToQuery(intra int, cross int) {
	if stub.SetNumPeersToQueryCalled != nil {
		stub.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (stub *RequesterStub) NumPeersToQuery() (int, int) {
	if stub.NumPeersToQueryCalled != nil {
		return stub.NumPeersToQueryCalled()
	}
	return 0, 0
}

// SetResolverDebugHandler -
func (stub *RequesterStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if stub.SetResolverDebugHandlerCalled != nil {
		return stub.SetResolverDebugHandlerCalled(handler)
	}
	return nil
}

// IsInterfaceNil -
func (stub *RequesterStub) IsInterfaceNil() bool {
	return stub == nil
}
