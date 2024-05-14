package systemSmartContracts

import "github.com/multiversx/mx-chain-go/vm"

type systemVMEEICreator struct {
}

// NewSystemVMEEICreator creates a system vm eei creator
func NewSystemVMEEICreator() *systemVMEEICreator {
	return &systemVMEEICreator{}
}

// CreateVmContext creates vm context for sovereign chain
func (vcc *systemVMEEICreator) CreateVmContext(args VMContextArgs) (vm.ContextHandler, error) {
	vmCtxt, err := NewVMContext(args)
	if err != nil {
		return nil, err
	}

	return NewSystemVMEEI(vmCtxt)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (vcc *systemVMEEICreator) IsInterfaceNil() bool {
	return vcc == nil
}
