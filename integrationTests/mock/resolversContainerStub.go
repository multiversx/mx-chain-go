package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ResolversContainerStub -
type ResolversContainerStub struct {
	GetCalled          func(key string) (dataRetriever.Resolver, error)
	AddCalled          func(key string, val dataRetriever.Resolver) error
	ReplaceCalled      func(key string, val dataRetriever.Resolver) error
	RemoveCalled       func(key string)
	LenCalled          func() int
	ResolverKeysCalled func() string
	IterateCalled      func(handler func(key string, resolver dataRetriever.Resolver) bool)
	CloseCalled        func() error
}

// Get -
func (rcs *ResolversContainerStub) Get(key string) (dataRetriever.Resolver, error) {
	return rcs.GetCalled(key)
}

// Add -
func (rcs *ResolversContainerStub) Add(key string, val dataRetriever.Resolver) error {
	return rcs.AddCalled(key, val)
}

// AddMultiple -
func (rcs *ResolversContainerStub) AddMultiple(_ []string, _ []dataRetriever.Resolver) error {
	panic("implement me")
}

// Replace -
func (rcs *ResolversContainerStub) Replace(key string, val dataRetriever.Resolver) error {
	return rcs.ReplaceCalled(key, val)
}

// Remove -
func (rcs *ResolversContainerStub) Remove(key string) {
	rcs.RemoveCalled(key)
}

// Len -
func (rcs *ResolversContainerStub) Len() int {
	return rcs.LenCalled()
}

// ResolverKeys -
func (rcs *ResolversContainerStub) ResolverKeys() string {
	if rcs.ResolverKeysCalled != nil {
		return rcs.ResolverKeysCalled()
	}

	return ""
}

// Iterate -
func (rcs *ResolversContainerStub) Iterate(handler func(key string, resolver dataRetriever.Resolver) bool) {
	if rcs.IterateCalled != nil {
		rcs.IterateCalled(handler)
	}
}

// Close -
func (rcs *ResolversContainerStub) Close() error {
	if rcs.CloseCalled != nil {
		return rcs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcs *ResolversContainerStub) IsInterfaceNil() bool {
	return rcs == nil
}
