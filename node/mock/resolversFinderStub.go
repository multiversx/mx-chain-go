package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type ResolversFinderStub struct {
	GetCalled                func(key string) (process.Resolver, error)
	AddCalled                func(key string, val process.Resolver) error
	ReplaceCalled            func(key string, val process.Resolver) error
	RemoveCalled             func(key string)
	LenCalled                func() int
	IntraShardResolverCalled func(baseTopic string) (process.Resolver, error)
	CrossShardResolverCalled func(baseTopic string, crossShard uint32) (process.Resolver, error)
}

func (rfs *ResolversFinderStub) Get(key string) (process.Resolver, error) {
	return rfs.GetCalled(key)
}

func (rfs *ResolversFinderStub) Add(key string, val process.Resolver) error {
	return rfs.AddCalled(key, val)
}

func (rfs *ResolversFinderStub) AddMultiple(keys []string, resolvers []process.Resolver) error {
	panic("implement me")
}

func (rfs *ResolversFinderStub) Replace(key string, val process.Resolver) error {
	return rfs.ReplaceCalled(key, val)
}

func (rfs *ResolversFinderStub) Remove(key string) {
	rfs.RemoveCalled(key)
}

func (rfs *ResolversFinderStub) Len() int {
	return rfs.LenCalled()
}

func (rfs *ResolversFinderStub) IntraShardResolver(baseTopic string) (process.Resolver, error) {
	return rfs.IntraShardResolverCalled(baseTopic)
}

func (rfs *ResolversFinderStub) CrossShardResolver(baseTopic string, crossShard uint32) (process.Resolver, error) {
	return rfs.CrossShardResolverCalled(baseTopic, crossShard)
}
