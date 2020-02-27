package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorsContainerStub -
type InterceptorsContainerStub struct {
}

// Get -
func (ics *InterceptorsContainerStub) Get(_ string) (process.Interceptor, error) {
	panic("implement me")
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
