package testscommon

import (
	"github.com/multiversx/mx-chain-go/common"
)

// BlockInfoProviderStub -
type BlockInfoProviderStub struct {
	GetBlockInfoCalled func() common.BlockInfo
}

// GetBlockInfo -
func (stub *BlockInfoProviderStub) GetBlockInfo() common.BlockInfo {
	if stub.GetBlockInfoCalled != nil {
		return stub.GetBlockInfoCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *BlockInfoProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
