package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ResolversFinderStub -
type ResolversFinderStub struct {
	ResolversContainerStub
	IntraShardResolverCalled     func(baseTopic string) (dataRetriever.Resolver, error)
	MetaChainResolverCalled      func(baseTopic string) (dataRetriever.Resolver, error)
	CrossShardResolverCalled     func(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error)
	MetaCrossShardResolverCalled func(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error)
}

// MetaCrossShardResolver -
func (rfs *ResolversFinderStub) MetaCrossShardResolver(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
	return rfs.MetaCrossShardResolverCalled(baseTopic, crossShard)
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
