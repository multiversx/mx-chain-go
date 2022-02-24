package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// BlockChainHookHandlerStub -
type BlockChainHookHandlerStub struct {
	SetCurrentHeaderCalled             func(hdr data.HeaderHandler)
	NewAddressCalled                   func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	IsPayableCalled                    func(sndAddress []byte, rcvAddress []byte) (bool, error)
	DeleteCompiledCodeCalled           func(codeHash []byte)
	ProcessBuiltInFunctionCalled       func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
	FilterCodeMetadataForUpgradeCalled func(input []byte) ([]byte, int)
	ApplyFiltersOnCodeMetadataCalled   func(codeMetadata vmcommon.CodeMetadata) vmcommon.CodeMetadata
}

// ProcessBuiltInFunction -
func (e *BlockChainHookHandlerStub) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if e.ProcessBuiltInFunctionCalled != nil {
		return e.ProcessBuiltInFunctionCalled(input)
	}
	return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
}

// IsPayable -
func (e *BlockChainHookHandlerStub) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	if e.IsPayableCalled != nil {
		return e.IsPayableCalled(sndAddress, recvAddress)
	}
	return true, nil
}

// SetCurrentHeader -
func (e *BlockChainHookHandlerStub) SetCurrentHeader(hdr data.HeaderHandler) {
	if e.SetCurrentHeaderCalled != nil {
		e.SetCurrentHeaderCalled(hdr)
	}
}

// SaveNFTMetaDataToSystemAccount -
func (e *BlockChainHookHandlerStub) SaveNFTMetaDataToSystemAccount(_ data.TransactionHandler) error {
	return nil
}

// NewAddress -
func (e *BlockChainHookHandlerStub) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if e.NewAddressCalled != nil {
		return e.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}

	return make([]byte, 0), nil
}

// DeleteCompiledCode -
func (e *BlockChainHookHandlerStub) DeleteCompiledCode(codeHash []byte) {
	if e.DeleteCompiledCodeCalled != nil {
		e.DeleteCompiledCodeCalled(codeHash)
	}
}

// FilterCodeMetadataForUpgrade -
func (e *BlockChainHookHandlerStub) FilterCodeMetadataForUpgrade(input []byte) ([]byte, int) {
	if e.FilterCodeMetadataForUpgradeCalled != nil {
		return e.FilterCodeMetadataForUpgradeCalled(input)
	}

	return input, 0
}

// ApplyFiltersOnCodeMetadata -
func (e *BlockChainHookHandlerStub) ApplyFiltersOnCodeMetadata(codeMetadata vmcommon.CodeMetadata) vmcommon.CodeMetadata {
	if e.ApplyFiltersOnCodeMetadataCalled != nil {
		return e.ApplyFiltersOnCodeMetadataCalled(codeMetadata)
	}

	return codeMetadata
}

// IsInterfaceNil -
func (e *BlockChainHookHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
