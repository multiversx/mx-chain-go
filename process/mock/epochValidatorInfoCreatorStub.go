package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/state"
)

// EpochValidatorInfoCreatorStub -
type EpochValidatorInfoCreatorStub struct {
	CreateValidatorInfoMiniBlocksCalled           func(validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyValidatorInfoMiniBlocksCalled           func(miniblocks []*block.MiniBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error
	CreateMarshalizedDataCalled                   func(body block.Body) map[string][][]byte
	SaveValidatorInfoBlockDataToStorageCalled     func(metaBlock data.HeaderHandler, body *block.Body)
	DeleteValidatorInfoBlockDataFromStorageCalled func(metaBlock data.HeaderHandler, body *block.Body)
	RemoveBlockDataFromPoolsCalled                func(metaBlock data.HeaderHandler, body *block.Body)
}

// CreateValidatorInfoMiniBlocks -
func (e *EpochValidatorInfoCreatorStub) CreateValidatorInfoMiniBlocks(validatorInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if e.CreateValidatorInfoMiniBlocksCalled != nil {
		return e.CreateValidatorInfoMiniBlocksCalled(validatorInfo)
	}
	return make(block.MiniBlockSlice, 0), nil
}

// VerifyValidatorInfoMiniBlocks -
func (e *EpochValidatorInfoCreatorStub) VerifyValidatorInfoMiniBlocks(miniblocks []*block.MiniBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	if e.VerifyValidatorInfoMiniBlocksCalled != nil {
		return e.VerifyValidatorInfoMiniBlocksCalled(miniblocks, validatorsInfo)
	}
	return nil
}

// SaveValidatorInfoBlockDataToStorage -
func (e *EpochValidatorInfoCreatorStub) SaveValidatorInfoBlockDataToStorage(metaBlock data.HeaderHandler, body *block.Body) {
	if e.SaveValidatorInfoBlockDataToStorageCalled != nil {
		e.SaveValidatorInfoBlockDataToStorageCalled(metaBlock, body)
	}
}

// DeleteValidatorInfoBlockDataFromStorage -
func (e *EpochValidatorInfoCreatorStub) DeleteValidatorInfoBlockDataFromStorage(metaBlock data.HeaderHandler, body *block.Body) {
	if e.DeleteValidatorInfoBlockDataFromStorageCalled != nil {
		e.DeleteValidatorInfoBlockDataFromStorageCalled(metaBlock, body)
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
