package systemSmartContracts

import "github.com/multiversx/mx-chain-go/vm"

type oneShardSystemVMEEICreator struct {
}

// NewOneShardSystemVMEEICreator creates a system vm eei creator
func NewOneShardSystemVMEEICreator() *oneShardSystemVMEEICreator {
	return &oneShardSystemVMEEICreator{}
}

// CreateVmContext creates vm context for single shard chains
func (vcc *oneShardSystemVMEEICreator) CreateVmContext(args VMContextArgs) (vm.ContextHandler, error) {
	vmCtxt, err := NewVMContext(args)
	if err != nil {
		return nil, err
	}

	return NewOneShardSystemVMEEI(vmCtxt)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (vcc *oneShardSystemVMEEICreator) IsInterfaceNil() bool {
	return vcc == nil
}
