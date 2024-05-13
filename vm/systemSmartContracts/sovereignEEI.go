package systemSmartContracts

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/vm"
)

type sovereignVMContext struct {
	vm.ContextHandler
}

// NewSovereignVMContext creates a new sovereign vm context
func NewSovereignVMContext(vmContext vm.ContextHandler) (*sovereignVMContext, error) {
	if check.IfNil(vmContext) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	return &sovereignVMContext{
		vmContext,
	}, nil
}

// SendGlobalSettingToAll handles sending global settings information
func (sovHost *sovereignVMContext) SendGlobalSettingToAll(sender []byte, input []byte) error {
	return sovHost.ProcessBuiltInFunction(core.SystemAccountAddress, sender, big.NewInt(0), input, 0)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sovHost *sovereignVMContext) IsInterfaceNil() bool {
	return sovHost == nil
}
