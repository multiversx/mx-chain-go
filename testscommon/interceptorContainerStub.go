package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// InterceptorsContainerStub -
type InterceptorsContainerStub struct {
	IterateCalled     func(handler func(key string, interceptor core.Interceptor) bool)
	GetCalled         func(string) (core.Interceptor, error)
	AddCalled         func(key string, interceptor core.Interceptor) error
	AddMultipleCalled func(keys []string, interceptors []core.Interceptor) error
	ReplaceCalled     func(key string, interceptor core.Interceptor) error
	RemoveCalled      func(key string)
	LenCalled         func() int
	CloseCalled       func() error
}

// Iterate -
func (ics *InterceptorsContainerStub) Iterate(handler func(key string, interceptor core.Interceptor) bool) {
	if ics.IterateCalled != nil {
		ics.IterateCalled(handler)
	}
}

// Get -
func (ics *InterceptorsContainerStub) Get(topic string) (core.Interceptor, error) {
	if ics.GetCalled != nil {
		return ics.GetCalled(topic)
	}

	return &InterceptorStub{
		ProcessReceivedMessageCalled: func(message core.MessageP2P) error {
			return nil
		},
	}, nil
}

// Add -
func (ics *InterceptorsContainerStub) Add(key string, interceptor core.Interceptor) error {
	if ics.AddCalled != nil {
		return ics.AddCalled(key, interceptor)
	}

	return nil
}

// AddMultiple -
func (ics *InterceptorsContainerStub) AddMultiple(keys []string, interceptors []core.Interceptor) error {
	if ics.AddMultipleCalled != nil {
		return ics.AddMultipleCalled(keys, interceptors)
	}

	return nil
}

// Replace -
func (ics *InterceptorsContainerStub) Replace(key string, interceptor core.Interceptor) error {
	if ics.ReplaceCalled != nil {
		return ics.ReplaceCalled(key, interceptor)
	}

	return nil
}

// Remove -
func (ics *InterceptorsContainerStub) Remove(key string) {
	if ics.RemoveCalled != nil {
		ics.RemoveCalled(key)
	}
}

// Len -
func (ics *InterceptorsContainerStub) Len() int {
	if ics.LenCalled != nil {
		return ics.LenCalled()
	}

	return 0
}

// Close -
func (ics *InterceptorsContainerStub) Close() error {
	if ics.CloseCalled != nil {
		return ics.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (ics *InterceptorsContainerStub) IsInterfaceNil() bool {
	return ics == nil
}
