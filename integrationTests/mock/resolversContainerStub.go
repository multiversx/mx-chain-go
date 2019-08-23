package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type ResolversContainerStub struct {
	GetCalled     func(key string) (dataRetriever.Resolver, error)
	AddCalled     func(key string, val dataRetriever.Resolver) error
	ReplaceCalled func(key string, val dataRetriever.Resolver) error
	RemoveCalled  func(key string)
	LenCalled     func() int
}

func (rcs *ResolversContainerStub) Get(key string) (dataRetriever.Resolver, error) {
	return rcs.GetCalled(key)
}

func (rcs *ResolversContainerStub) Add(key string, val dataRetriever.Resolver) error {
	return rcs.AddCalled(key, val)
}

func (rcs *ResolversContainerStub) AddMultiple(keys []string, resolvers []dataRetriever.Resolver) error {
	panic("implement me")
}

func (rcs *ResolversContainerStub) Replace(key string, val dataRetriever.Resolver) error {
	return rcs.ReplaceCalled(key, val)
}

func (rcs *ResolversContainerStub) Remove(key string) {
	rcs.RemoveCalled(key)
}

func (rcs *ResolversContainerStub) Len() int {
	return rcs.LenCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcs *ResolversContainerStub) IsInterfaceNil() bool {
	if rcs == nil {
		return true
	}
	return false
}
