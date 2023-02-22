package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// RequestersContainerStub -
type RequestersContainerStub struct {
	GetCalled           func(key string) (dataRetriever.Requester, error)
	AddCalled           func(key string, val dataRetriever.Requester) error
	ReplaceCalled       func(key string, val dataRetriever.Requester) error
	RemoveCalled        func(key string)
	LenCalled           func() int
	RequesterKeysCalled func() string
	IterateCalled       func(handler func(key string, requester dataRetriever.Requester) bool)
	CloseCalled         func() error
}

// Get -
func (rcs *RequestersContainerStub) Get(key string) (dataRetriever.Requester, error) {
	return rcs.GetCalled(key)
}

// Add -
func (rcs *RequestersContainerStub) Add(key string, val dataRetriever.Requester) error {
	return rcs.AddCalled(key, val)
}

// AddMultiple -
func (rcs *RequestersContainerStub) AddMultiple(_ []string, _ []dataRetriever.Requester) error {
	return nil
}

// Replace -
func (rcs *RequestersContainerStub) Replace(key string, val dataRetriever.Requester) error {
	return rcs.ReplaceCalled(key, val)
}

// Remove -
func (rcs *RequestersContainerStub) Remove(key string) {
	rcs.RemoveCalled(key)
}

// Len -
func (rcs *RequestersContainerStub) Len() int {
	return rcs.LenCalled()
}

// RequesterKeys -
func (rcs *RequestersContainerStub) RequesterKeys() string {
	if rcs.RequesterKeysCalled != nil {
		return rcs.RequesterKeysCalled()
	}

	return ""
}

// Iterate -
func (rcs *RequestersContainerStub) Iterate(handler func(key string, requester dataRetriever.Requester) bool) {
	if rcs.IterateCalled != nil {
		rcs.IterateCalled(handler)
	}
}

// Close -
func (rcs *RequestersContainerStub) Close() error {
	if rcs.CloseCalled != nil {
		return rcs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcs *RequestersContainerStub) IsInterfaceNil() bool {
	return rcs == nil
}
