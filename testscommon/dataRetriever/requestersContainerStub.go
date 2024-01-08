package dataRetriever

import "github.com/multiversx/mx-chain-go/dataRetriever"

// RequestersContainerStub -
type RequestersContainerStub struct {
	GetCalled           func(key string) (dataRetriever.Requester, error)
	AddCalled           func(key string, val dataRetriever.Requester) error
	AddMultipleCalled   func(keys []string, requesters []dataRetriever.Requester) error
	ReplaceCalled       func(key string, val dataRetriever.Requester) error
	RemoveCalled        func(key string)
	LenCalled           func() int
	RequesterKeysCalled func() string
	IterateCalled       func(handler func(key string, requester dataRetriever.Requester) bool)
	CloseCalled         func() error
}

// Get -
func (stub *RequestersContainerStub) Get(key string) (dataRetriever.Requester, error) {
	if stub.GetCalled != nil {
		return stub.GetCalled(key)
	}
	return nil, nil
}

// Add -
func (stub *RequestersContainerStub) Add(key string, val dataRetriever.Requester) error {
	if stub.AddCalled != nil {
		return stub.AddCalled(key, val)
	}
	return nil
}

// AddMultiple -
func (stub *RequestersContainerStub) AddMultiple(keys []string, requesters []dataRetriever.Requester) error {
	if stub.AddMultipleCalled != nil {
		return stub.AddMultipleCalled(keys, requesters)
	}
	return nil
}

// Replace -
func (stub *RequestersContainerStub) Replace(key string, val dataRetriever.Requester) error {
	if stub.ReplaceCalled != nil {
		return stub.ReplaceCalled(key, val)
	}
	return nil
}

// Remove -
func (stub *RequestersContainerStub) Remove(key string) {
	if stub.RemoveCalled != nil {
		stub.RemoveCalled(key)
	}
}

// Len -
func (stub *RequestersContainerStub) Len() int {
	if stub.LenCalled != nil {
		return stub.LenCalled()
	}
	return 0
}

// RequesterKeys -
func (stub *RequestersContainerStub) RequesterKeys() string {
	if stub.RequesterKeysCalled != nil {
		return stub.RequesterKeysCalled()
	}
	return ""
}

// Iterate -
func (stub *RequestersContainerStub) Iterate(handler func(key string, requester dataRetriever.Requester) bool) {
	if stub.IterateCalled != nil {
		stub.IterateCalled(handler)
	}
}

// Close -
func (stub *RequestersContainerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *RequestersContainerStub) IsInterfaceNil() bool {
	return stub == nil
}
