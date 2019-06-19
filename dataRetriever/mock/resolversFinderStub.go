package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type ResolversFinderStub struct {
	ResolversContainerStub
	IntraShardResolverCalled func(baseTopic string) (dataRetriever.Resolver, error)
	MetaChainResolverCalled  func(baseTopic string) (dataRetriever.Resolver, error)
	CrossShardResolverCalled func(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error)
}

func (rfs *ResolversFinderStub) IntraShardResolver(baseTopic string) (dataRetriever.Resolver, error) {
	return rfs.IntraShardResolverCalled(baseTopic)
}

func (rfs *ResolversFinderStub) MetaChainResolver(baseTopic string) (dataRetriever.Resolver, error) {
	return rfs.MetaChainResolverCalled(baseTopic)
}

func (rfs *ResolversFinderStub) CrossShardResolver(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
	return rfs.CrossShardResolverCalled(baseTopic, crossShard)
}
