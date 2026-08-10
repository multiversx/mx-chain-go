package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// EpochStartDataCreatorStub -
type EpochStartDataCreatorStub struct {
	CreateEpochStartDataCalled                 func() (*block.EpochStart, error)
	CreateEpochStartShardDataMetablockV3Called func(metablock data.MetaHeaderHandler) ([]block.EpochStartShardData, error)
	VerifyEpochStartDataForMetablockCalled     func(metaBlock data.MetaHeaderHandler) error
}

// CreateEpochStartData -
func (e *EpochStartDataCreatorStub) CreateEpochStartData() (*block.EpochStart, error) {
	if e.CreateEpochStartDataCalled != nil {
		return e.CreateEpochStartDataCalled()
	}
	return &block.EpochStart{
		LastFinalizedHeaders: []block.EpochStartShardData{{}},
		Economics: block.Economics{
			RewardsForProtocolSustainability: big.NewInt(0)},
	}, nil
}

// CreateEpochStartShardDataMetablockV3 -
func (e *EpochStartDataCreatorStub) CreateEpochStartShardDataMetablockV3(metablock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
	if e.CreateEpochStartShardDataMetablockV3Called != nil {
		return e.CreateEpochStartShardDataMetablockV3Called(metablock)
	}
	return []block.EpochStartShardData{{}}, nil
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
