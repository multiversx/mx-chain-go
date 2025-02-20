package systemSmartContracts

import "github.com/multiversx/mx-chain-go/vm"

type vmContextCreator struct {
}

// NewVMContextCreator creates a vm context creator
func NewVMContextCreator() *vmContextCreator {
	return &vmContextCreator{}
}

// CreateVmContext creates vm context for regular chain
func (vcc *vmContextCreator) CreateVmContext(args VMContextArgs) (vm.ContextHandler, error) {
	return NewVMContext(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (vcc *vmContextCreator) IsInterfaceNil() bool {
	return vcc == nil
}
