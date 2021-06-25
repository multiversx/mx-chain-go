package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// BlockChainHookHandlerMock -
type BlockChainHookHandlerMock struct {
	SetCurrentHeaderCalled       func(hdr data.HeaderHandler)
	NewAddressCalled             func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	IsPayableCalled              func(address []byte) (bool, error)
	DeleteCompiledCodeCalled     func(codeHash []byte)
	ProcessBuiltInFunctionCalled func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// ProcessBuiltInFunction -
func (e *BlockChainHookHandlerMock) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if e.ProcessBuiltInFunctionCalled != nil {
		return e.ProcessBuiltInFunctionCalled(input)
	}
	return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
}

// IsPayable -
func (e *BlockChainHookHandlerMock) IsPayable(address []byte) (bool, error) {
	if e.IsPayableCalled != nil {
		return e.IsPayableCalled(address)
	}
	return true, nil
}

// SetCurrentHeader -
func (e *BlockChainHookHandlerMock) SetCurrentHeader(hdr data.HeaderHandler) {
	if e.SetCurrentHeaderCalled != nil {
		e.SetCurrentHeaderCalled(hdr)
	}
}

// NewAddress -
func (e *BlockChainHookHandlerMock) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if e.NewAddressCalled != nil {
		return e.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}

	return make([]byte, 0), nil
}

// DeleteCompiledCode -
func (e *BlockChainHookHandlerMock) DeleteCompiledCode(codeHash []byte) {
	if e.DeleteCompiledCodeCalled != nil {
		e.DeleteCompiledCodeCalled(codeHash)
	}
}

// IsInterfaceNil -
func (e *BlockChainHookHandlerMock) IsInterfaceNil() bool {
	return e == nil
}
