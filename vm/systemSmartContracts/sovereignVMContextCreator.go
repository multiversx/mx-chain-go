package systemSmartContracts

import "github.com/multiversx/mx-chain-go/vm"

type sovereignVMContextCreator struct {
}

// NewSovereignVMContextCreator creates a sovereign vm context creator
func NewSovereignVMContextCreator() *sovereignVMContextCreator {
	return &sovereignVMContextCreator{}
}

// CreateVmContext creates vm context for sovereign chain
func (vcc *sovereignVMContextCreator) CreateVmContext(args VMContextArgs) (vm.ContextHandler, error) {
	vmCtxt, err := NewVMContext(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignVMContext(vmCtxt)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (vcc *sovereignVMContextCreator) IsInterfaceNil() bool {
	return vcc == nil
}
