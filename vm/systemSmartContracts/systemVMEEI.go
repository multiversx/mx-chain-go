package systemSmartContracts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/vm"
)

type systemVMEEI struct {
	vm.ContextHandler
}

// NewSystemVMEEI creates a new system vm eei context, mainly used for now in sovereign chain
func NewSystemVMEEI(vmContext vm.ContextHandler) (*systemVMEEI, error) {
	if check.IfNil(vmContext) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	return &systemVMEEI{
		vmContext,
	}, nil
}

// SendGlobalSettingToAll handles sending global settings information
func (sovHost *systemVMEEI) SendGlobalSettingToAll(sender []byte, input []byte) error {
	return sovHost.ProcessBuiltInFunction(core.SystemAccountAddress, sender, big.NewInt(0), input, 0)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sovHost *systemVMEEI) IsInterfaceNil() bool {
	return sovHost == nil
}
