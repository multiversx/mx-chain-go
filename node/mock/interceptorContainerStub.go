package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type InterceptorsContainerStub struct {
}

func (ics *InterceptorsContainerStub) Get(key string) (process.Interceptor, error) {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Add(key string, val process.Interceptor) error {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) AddMultiple(keys []string, interceptors []process.Interceptor) error {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Replace(key string, val process.Interceptor) error {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Remove(key string) {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Len() int {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ics *InterceptorsContainerStub) IsInterfaceNil() bool {
	if ics == nil {
		return true
	}
	return false
}
