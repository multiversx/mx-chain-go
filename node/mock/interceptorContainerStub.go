package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type InterceptorsContainerStub struct {
}

func (ics *InterceptorsContainerStub) Get(_ string) (process.Interceptor, error) {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Add(_ string, _ process.Interceptor) error {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) AddMultiple(_ []string, _ []process.Interceptor) error {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Replace(_ string, _ process.Interceptor) error {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Remove(_ string) {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Len() int {
	panic("implement me")
}

func (ics *InterceptorsContainerStub) Iterate(_ func(key string, interceptor process.Interceptor) bool) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ics *InterceptorsContainerStub) IsInterfaceNil() bool {
	return ics == nil
}
