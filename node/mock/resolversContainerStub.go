package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type ResolversContainerStub struct {
	GetCalled     func(key string) (process.Resolver, error)
	AddCalled     func(key string, val process.Resolver) error
	ReplaceCalled func(key string, val process.Resolver) error
	RemoveCalled  func(key string)
	LenCalled     func() int
}

func (rcs *ResolversContainerStub) Get(key string) (process.Resolver, error) {
	return rcs.GetCalled(key)
}

func (rcs *ResolversContainerStub) Add(key string, val process.Resolver) error {
	return rcs.AddCalled(key, val)
}

func (rcs *ResolversContainerStub) Replace(key string, val process.Resolver) error {
	return rcs.ReplaceCalled(key, val)
}

func (rcs *ResolversContainerStub) Remove(key string) {
	rcs.RemoveCalled(key)
}

func (rcs *ResolversContainerStub) Len() int {
	return rcs.LenCalled()
}
