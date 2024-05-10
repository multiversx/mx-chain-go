package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

// EpochValidatorInfoCreatorStub -
type EpochValidatorInfoCreatorStub struct {
	CreateValidatorInfoMiniBlocksCalled func(validatorsInfo state.ShardValidatorsInfoMapHandler) (block.MiniBlockSlice, error)
	VerifyValidatorInfoMiniBlocksCalled func(miniblocks []*block.MiniBlock, validatorsInfo state.ShardValidatorsInfoMapHandler) error
	GetLocalValidatorInfoCacheCalled    func() epochStart.ValidatorInfoCacher
	CreateMarshalledDataCalled          func(body *block.Body) map[string][][]byte
	GetValidatorInfoTxsCalled           func(body *block.Body) map[string]*state.ShardValidatorInfo
	SaveBlockDataToStorageCalled        func(metaBlock data.HeaderHandler, body *block.Body)
	DeleteBlockDataFromStorageCalled    func(metaBlock data.HeaderHandler, body *block.Body)
	RemoveBlockDataFromPoolsCalled      func(metaBlock data.HeaderHandler, body *block.Body)
}

// CreateValidatorInfoMiniBlocks -
func (e *EpochValidatorInfoCreatorStub) CreateValidatorInfoMiniBlocks(validatorInfo state.ShardValidatorsInfoMapHandler) (block.MiniBlockSlice, error) {
	if e.CreateValidatorInfoMiniBlocksCalled != nil {
		return e.CreateValidatorInfoMiniBlocksCalled(validatorInfo)
	}
	return make(block.MiniBlockSlice, 0), nil
}

// VerifyValidatorInfoMiniBlocks -
func (e *EpochValidatorInfoCreatorStub) VerifyValidatorInfoMiniBlocks(miniBlocks []*block.MiniBlock, validatorsInfo state.ShardValidatorsInfoMapHandler) error {
	if e.VerifyValidatorInfoMiniBlocksCalled != nil {
		return e.VerifyValidatorInfoMiniBlocksCalled(miniBlocks, validatorsInfo)
	}
	return nil
}

// GetLocalValidatorInfoCache -
func (e *EpochValidatorInfoCreatorStub) GetLocalValidatorInfoCache() epochStart.ValidatorInfoCacher {
	if e.GetLocalValidatorInfoCacheCalled != nil {
		return e.GetLocalValidatorInfoCacheCalled()
	}
	return nil
}

// CreateMarshalledData -
func (e *EpochValidatorInfoCreatorStub) CreateMarshalledData(body *block.Body) map[string][][]byte {
	if e.CreateMarshalledDataCalled != nil {
		return e.CreateMarshalledDataCalled(body)
	}
	return nil
}

// GetValidatorInfoTxs -
func (e *EpochValidatorInfoCreatorStub) GetValidatorInfoTxs(body *block.Body) map[string]*state.ShardValidatorInfo {
	if e.GetValidatorInfoTxsCalled != nil {
		return e.GetValidatorInfoTxsCalled(body)
	}
	return nil
}

// SaveBlockDataToStorage -
func (e *EpochValidatorInfoCreatorStub) SaveBlockDataToStorage(metaBlock data.HeaderHandler, body *block.Body) {
	if e.SaveBlockDataToStorageCalled != nil {
		e.SaveBlockDataToStorageCalled(metaBlock, body)
	}
}

// DeleteBlockDataFromStorage -
func (e *EpochValidatorInfoCreatorStub) DeleteBlockDataFromStorage(metaBlock data.HeaderHandler, body *block.Body) {
	if e.DeleteBlockDataFromStorageCalled != nil {
		e.DeleteBlockDataFromStorageCalled(metaBlock, body)
	}
}

// IsInterfaceNil -
func (e *EpochValidatorInfoCreatorStub) IsInterfaceNil() bool {
	return e == nil
}

// RemoveBlockDataFromPools -
func (e *EpochValidatorInfoCreatorStub) RemoveBlockDataFromPools(metaBlock data.HeaderHandler, body *block.Body) {
	if e.RemoveBlockDataFromPoolsCalled != nil {
		e.RemoveBlockDataFromPoolsCalled(metaBlock, body)
	}
}
