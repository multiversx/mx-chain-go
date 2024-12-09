package systemSmartContracts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/vm"
)

type oneShardSystemVMEEI struct {
	*vmContext
}

// NewOneShardSystemVMEEI creates a new system vm eei context, mainly used for now in sovereign chain
func NewOneShardSystemVMEEI(vmContext *vmContext) (*oneShardSystemVMEEI, error) {
	if check.IfNil(vmContext) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	return &oneShardSystemVMEEI{
		vmContext,
	}, nil
}

// SendGlobalSettingToAll handles sending global settings information
func (host *oneShardSystemVMEEI) SendGlobalSettingToAll(sender []byte, input []byte) error {
	return host.ProcessBuiltInFunction(core.SystemAccountAddress, sender, big.NewInt(0), input, 0)
}

// ProcessBuiltInFunction will execute process if sender and destination is same shard/sovereign
func (host *oneShardSystemVMEEI) ProcessBuiltInFunction(
	destination []byte,
	sender []byte,
	value *big.Int,
	input []byte,
	gasLimit uint64,
) error {
	host.Transfer(destination, sender, value, input, gasLimit)
	return host.processBuiltInFunction(destination, sender, value, input, gasLimit)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (host *oneShardSystemVMEEI) IsInterfaceNil() bool {
	return host == nil
}
