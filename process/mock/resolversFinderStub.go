package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type ResolversFinderStub struct {
	ResolversContainerStub
	IntraShardResolverCalled func(baseTopic string) (process.Resolver, error)
	CrossShardResolverCalled func(baseTopic string, crossShard uint32) (process.Resolver, error)
}

func (rfs *ResolversFinderStub) IntraShardResolver(baseTopic string) (process.Resolver, error) {
	return rfs.IntraShardResolverCalled(baseTopic)
}

func (rfs *ResolversFinderStub) CrossShardResolver(baseTopic string, crossShard uint32) (process.Resolver, error) {
	return rfs.CrossShardResolverCalled(baseTopic, crossShard)
}
