package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// ResolversContainerStub -
type ResolversContainerStub struct {
	GetCalled     func(key string) (dataRetriever.Resolver, error)
	AddCalled     func(key string, val dataRetriever.Resolver) error
	ReplaceCalled func(key string, val dataRetriever.Resolver) error
	RemoveCalled  func(key string)
	LenCalled     func() int
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

// IsInterfaceNil returns true if there is no value under the interface
func (rcs *ResolversContainerStub) IsInterfaceNil() bool {
	return rcs == nil
}
