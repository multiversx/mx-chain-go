package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ResolversFinderStub -
type ResolversFinderStub struct {
	GetCalled                    func(key string) (dataRetriever.Resolver, error)
	AddCalled                    func(key string, val dataRetriever.Resolver) error
	ReplaceCalled                func(key string, val dataRetriever.Resolver) error
	RemoveCalled                 func(key string)
	LenCalled                    func() int
	IntraShardResolverCalled     func(baseTopic string) (dataRetriever.Resolver, error)
	MetaChainResolverCalled      func(baseTopic string) (dataRetriever.Resolver, error)
	CrossShardResolverCalled     func(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error)
	ResolverKeysCalled           func() string
	MetaCrossShardResolverCalled func(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error)
	IterateCalled                func(handler func(key string, resolver dataRetriever.Resolver) bool)
}

// MetaCrossShardResolver -
func (rfs *ResolversFinderStub) MetaCrossShardResolver(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
	return rfs.MetaCrossShardResolverCalled(baseTopic, crossShard)
}

// Get -
func (rfs *ResolversFinderStub) Get(key string) (dataRetriever.Resolver, error) {
	return rfs.GetCalled(key)
}

// Add -
func (rfs *ResolversFinderStub) Add(key string, val dataRetriever.Resolver) error {
	return rfs.AddCalled(key, val)
}

// AddMultiple -
func (rfs *ResolversFinderStub) AddMultiple(_ []string, _ []dataRetriever.Resolver) error {
	panic("implement me")
}

// Replace -
func (rfs *ResolversFinderStub) Replace(key string, val dataRetriever.Resolver) error {
	return rfs.ReplaceCalled(key, val)
}

// Remove -
func (rfs *ResolversFinderStub) Remove(key string) {
	rfs.RemoveCalled(key)
}

// Len -
func (rfs *ResolversFinderStub) Len() int {
	return rfs.LenCalled()
}

// Iterate -
func (rfs *ResolversFinderStub) Iterate(handler func(key string, resolver dataRetriever.Resolver) bool) {
	if rfs.IterateCalled != nil {
		rfs.IterateCalled(handler)
	}
}

// ResolverKeys -
func (rfs *ResolversFinderStub) ResolverKeys() string {
	if rfs.ResolverKeysCalled != nil {
		return rfs.ResolverKeysCalled()
	}

	return ""
}

// IntraShardResolver -
func (rfs *ResolversFinderStub) IntraShardResolver(baseTopic string) (dataRetriever.Resolver, error) {
	return rfs.IntraShardResolverCalled(baseTopic)
}

// MetaChainResolver -
func (rfs *ResolversFinderStub) MetaChainResolver(baseTopic string) (dataRetriever.Resolver, error) {
	return rfs.MetaChainResolverCalled(baseTopic)
}

// CrossShardResolver -
func (rfs *ResolversFinderStub) CrossShardResolver(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
	return rfs.CrossShardResolverCalled(baseTopic, crossShard)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rfs *ResolversFinderStub) IsInterfaceNil() bool {
	return rfs == nil
}
