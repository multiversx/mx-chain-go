package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// EpochRewardsCreatorStub -
type EpochRewardsCreatorStub struct {
	CreateRewardsMiniBlocksCalled func(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocksCalled func(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) error
	CreateMarshalizedDataCalled   func(body block.Body) map[string][][]byte
	SaveTxBlockToStorageCalled    func(metaBlock *block.MetaBlock, body block.Body)
	DeleteTxsFromStorageCalled    func(metaBlock *block.MetaBlock, body block.Body)
}

// CreateRewardsMiniBlocks -
func (e *EpochRewardsCreatorStub) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if e.CreateRewardsMiniBlocksCalled != nil {
		return e.CreateRewardsMiniBlocksCalled(metaBlock, validatorInfos)
	}
	return make(block.MiniBlockSlice, 0), nil
}

// VerifyRewardsMiniBlocks -
func (e *EpochRewardsCreatorStub) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorInfos map[uint32][]*state.ValidatorInfo) error {
	if e.VerifyRewardsMiniBlocksCalled != nil {
		return e.VerifyRewardsMiniBlocksCalled(metaBlock, validatorInfos)
	}
	return nil
}

// CreateMarshalizedData -
func (e *EpochRewardsCreatorStub) CreateMarshalizedData(body block.Body) map[string][][]byte {
	if e.CreateMarshalizedDataCalled != nil {
		return e.CreateMarshalizedDataCalled(body)
	}
	return nil
}

// SaveTxBlockToStorage -
func (e *EpochRewardsCreatorStub) SaveTxBlockToStorage(metaBlock *block.MetaBlock, body block.Body) {
	if e.SaveTxBlockToStorageCalled != nil {
		e.SaveTxBlockToStorageCalled(metaBlock, body)
	}
}

// DeleteTxsFromStorage -
func (e *EpochRewardsCreatorStub) DeleteTxsFromStorage(metaBlock *block.MetaBlock, body block.Body) {
	if e.DeleteTxsFromStorageCalled != nil {
		e.DeleteTxsFromStorageCalled(metaBlock, body)
	}
}

// IsInterfaceNil -
func (e *EpochRewardsCreatorStub) IsInterfaceNil() bool {
	return e == nil
}
