package systemSmartContracts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/vm"
)

type oneShardSystemVMEEI struct {
	vm.ContextHandler
}

// NewOneShardSystemVMEEI creates a new system vm eei context, mainly used for now in sovereign chain
func NewOneShardSystemVMEEI(vmContext vm.ContextHandler) (*oneShardSystemVMEEI, error) {
	if check.IfNil(vmContext) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	return &oneShardSystemVMEEI{
		vmContext,
	}, nil
}

// SendGlobalSettingToAll handles sending global settings information
func (sovHost *oneShardSystemVMEEI) SendGlobalSettingToAll(sender []byte, input []byte) error {
	// TODO: MX-15470 remove sameShard check from ProcessBuiltInFunction
	return sovHost.ProcessBuiltInFunction(core.SystemAccountAddress, sender, big.NewInt(0), input, 0)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sovHost *oneShardSystemVMEEI) IsInterfaceNil() bool {
	return sovHost == nil
}
