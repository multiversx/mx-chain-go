package testscommon

import (
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptorsContainerStub -
type InterceptorsContainerStub struct {
	IterateCalled     func(handler func(key string, interceptor process.Interceptor) bool)
	GetCalled         func(string) (process.Interceptor, error)
	AddCalled         func(key string, interceptor process.Interceptor) error
	AddMultipleCalled func(keys []string, interceptors []process.Interceptor) error
	ReplaceCalled     func(key string, interceptor process.Interceptor) error
	RemoveCalled      func(key string)
	LenCalled         func() int
	CloseCalled       func() error
}

// Iterate -
func (ics *InterceptorsContainerStub) Iterate(handler func(key string, interceptor process.Interceptor) bool) {
	if ics.IterateCalled != nil {
		ics.IterateCalled(handler)
	}
}

// Get -
func (ics *InterceptorsContainerStub) Get(topic string) (process.Interceptor, error) {
	if ics.GetCalled != nil {
		return ics.GetCalled(topic)
	}

	return &InterceptorStub{
		ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
			return nil
		},
	}, nil
}

// Add -
func (ics *InterceptorsContainerStub) Add(key string, interceptor process.Interceptor) error {
	if ics.AddCalled != nil {
		return ics.AddCalled(key, interceptor)
	}

	return nil
}

// AddMultiple -
func (ics *InterceptorsContainerStub) AddMultiple(keys []string, interceptors []process.Interceptor) error {
	if ics.AddMultipleCalled != nil {
		return ics.AddMultipleCalled(keys, interceptors)
	}

	return nil
}

// Replace -
func (ics *InterceptorsContainerStub) Replace(key string, interceptor process.Interceptor) error {
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
