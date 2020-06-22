package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorsContainerStub -
type InterceptorsContainerStub struct {
	GetCalled func(string) (process.Interceptor, error)
}

// Iterate -
func (ics *InterceptorsContainerStub) Iterate(_ func(key string, interceptor process.Interceptor) bool) {
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
func (ics *InterceptorsContainerStub) Add(_ string, _ process.Interceptor) error {
	panic("implement me")
}

// AddMultiple -
func (ics *InterceptorsContainerStub) AddMultiple(_ []string, _ []process.Interceptor) error {
	panic("implement me")
}

// Replace -
func (ics *InterceptorsContainerStub) Replace(_ string, _ process.Interceptor) error {
	panic("implement me")
}

// Remove -
func (ics *InterceptorsContainerStub) Remove(_ string) {
	panic("implement me")
}

// Len -
func (ics *InterceptorsContainerStub) Len() int {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ics *InterceptorsContainerStub) IsInterfaceNil() bool {
	return ics == nil
}
