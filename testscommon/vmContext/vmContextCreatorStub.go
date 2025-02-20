package vmContext

import (
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

// VMContextCreatorStub -
type VMContextCreatorStub struct {
	CreateVmContextCalled func(args systemSmartContracts.VMContextArgs) (vm.ContextHandler, error)
}

// CreateVmContext -
func (stub *VMContextCreatorStub) CreateVmContext(args systemSmartContracts.VMContextArgs) (vm.ContextHandler, error) {
	if stub.CreateVmContextCalled != nil {
		return stub.CreateVmContextCalled(args)
	}

	return &mock.SystemEIStub{}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (stub *VMContextCreatorStub) IsInterfaceNil() bool {
	return stub == nil
}
