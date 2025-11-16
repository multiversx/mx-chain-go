package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// EpochStartDataCreatorStub -
type EpochStartDataCreatorStub struct {
	CreateEpochStartDataCalled             func() (*block.EpochStart, error)
	CreateEpochStartDataMetablockV3Called  func(metablock data.MetaHeaderHandler) ([]block.EpochStartShardData, error)
	VerifyEpochStartDataForMetablockCalled func(metaBlock data.MetaHeaderHandler) error
}

// CreateEpochStartData -
func (e *EpochStartDataCreatorStub) CreateEpochStartData() (*block.EpochStart, error) {
	if e.CreateEpochStartDataCalled != nil {
		return e.CreateEpochStartDataCalled()
	}
	return &block.EpochStart{}, nil
}

// CreateEpochStartShardDataMetablockV3 -
func (e *EpochStartDataCreatorStub) CreateEpochStartShardDataMetablockV3(metablock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
	if e.CreateEpochStartDataMetablockV3Called != nil {
		return e.CreateEpochStartDataMetablockV3Called(metablock)
	}
	return nil, nil
}

// VerifyEpochStartDataForMetablock -
func (e *EpochStartDataCreatorStub) VerifyEpochStartDataForMetablock(metaBlock data.MetaHeaderHandler) error {
	if e.VerifyEpochStartDataForMetablockCalled != nil {
		return e.VerifyEpochStartDataForMetablockCalled(metaBlock)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochStartDataCreatorStub) IsInterfaceNil() bool {
	return e == nil
}
