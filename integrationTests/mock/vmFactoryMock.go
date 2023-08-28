package mock

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// VmMachinesContainerFactoryMock -
type VmMachinesContainerFactoryMock struct {
}

// Create -
func (v *VmMachinesContainerFactoryMock) Create() (process.VirtualMachinesContainer, error) {
	return &VMContainerMock{}, nil
}

// Close -
func (v *VmMachinesContainerFactoryMock) Close() error {
	return nil
}

// BlockChainHookImpl -
func (v *VmMachinesContainerFactoryMock) BlockChainHookImpl() process.BlockChainHookWithAccountsAdapter {
	return &testscommon.BlockChainHookStub{}
}

// IsInterfaceNil -
func (v *VmMachinesContainerFactoryMock) IsInterfaceNil() bool {
	return v == nil
}
