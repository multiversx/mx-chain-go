package dataRetriever

import "github.com/multiversx/mx-chain-go/dataRetriever"

// RequestersFinderStub -
type RequestersFinderStub struct {
	IntraShardRequesterCalled     func(baseTopic string) (dataRetriever.Requester, error)
	MetaChainRequesterCalled      func(baseTopic string) (dataRetriever.Requester, error)
	CrossShardRequesterCalled     func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error)
	MetaCrossShardRequesterCalled func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error)
	GetCalled                     func(key string) (dataRetriever.Requester, error)
	AddCalled                     func(key string, val dataRetriever.Requester) error
	AddMultipleCalled             func(keys []string, requesters []dataRetriever.Requester) error
	ReplaceCalled                 func(key string, val dataRetriever.Requester) error
	RemoveCalled                  func(key string)
	LenCalled                     func() int
	RequesterKeysCalled           func() string
	IterateCalled                 func(handler func(key string, requester dataRetriever.Requester) bool)
	CloseCalled                   func() error
}

// IntraShardRequester -
func (stub *RequestersFinderStub) IntraShardRequester(baseTopic string) (dataRetriever.Requester, error) {
	if stub.IntraShardRequesterCalled != nil {
		return stub.IntraShardRequesterCalled(baseTopic)
	}
	return nil, nil
}

// MetaChainRequester -
func (stub *RequestersFinderStub) MetaChainRequester(baseTopic string) (dataRetriever.Requester, error) {
	if stub.MetaChainRequesterCalled != nil {
		return stub.MetaChainRequesterCalled(baseTopic)
	}
	return nil, nil
}

// CrossShardRequester -
func (stub *RequestersFinderStub) CrossShardRequester(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
	if stub.CrossShardRequesterCalled != nil {
		return stub.CrossShardRequesterCalled(baseTopic, crossShard)
	}
	return nil, nil
}

// MetaCrossShardRequester -
func (stub *RequestersFinderStub) MetaCrossShardRequester(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
	if stub.MetaCrossShardRequesterCalled != nil {
		return stub.MetaCrossShardRequesterCalled(baseTopic, crossShard)
	}
	return nil, nil
}

// Get -
func (stub *RequestersFinderStub) Get(key string) (dataRetriever.Requester, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(key)
	}
	return nil, nil
}

// Add -
func (stub *RequestersFinderStub) Add(key string, val dataRetriever.Requester) error {
	if stub.AddCalled != nil {
		return stub.AddCalled(key, val)
	}
	return nil
}

// AddMultiple -
func (stub *RequestersFinderStub) AddMultiple(keys []string, requesters []dataRetriever.Requester) error {
	if stub.AddMultipleCalled != nil {
		return stub.AddMultipleCalled(keys, requesters)
	}
	return nil
}

// Replace -
func (stub *RequestersFinderStub) Replace(key string, val dataRetriever.Requester) error {
	if stub.ReplaceCalled != nil {
		return stub.ReplaceCalled(key, val)
	}
	return nil
}

// Remove -
func (stub *RequestersFinderStub) Remove(key string) {
	if stub.RemoveCalled != nil {
		stub.RemoveCalled(key)
	}
}

// Len -
func (stub *RequestersFinderStub) Len() int {
	if stub.LenCalled != nil {
		return stub.LenCalled()
	}
	return 0
}

// RequesterKeys -
func (stub *RequestersFinderStub) RequesterKeys() string {
	if stub.RequesterKeysCalled != nil {
		return stub.RequesterKeysCalled()
	}
	return ""
}

// Iterate -
func (stub *RequestersFinderStub) Iterate(handler func(key string, requester dataRetriever.Requester) bool) {
	if stub.IterateCalled != nil {
		stub.IterateCalled(handler)
	}
}

// Close -
func (stub *RequestersFinderStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *RequestersFinderStub) IsInterfaceNil() bool {
	return stub == nil
}
