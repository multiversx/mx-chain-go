package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"math/big"
)

// EpochStartDataCreatorStub -
type EpochStartDataCreatorStub struct {
	CreateEpochStartDataCalled             func() (*block.EpochStart, error)
	VerifyEpochStartDataForMetablockCalled func(metaBlock *block.MetaBlock) error
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

// VerifyEpochStartDataForMetablock -
func (e *EpochStartDataCreatorStub) VerifyEpochStartDataForMetablock(metaBlock *block.MetaBlock) error {
	if e.VerifyEpochStartDataForMetablockCalled != nil {
		return e.VerifyEpochStartDataForMetablockCalled(metaBlock)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochStartDataCreatorStub) IsInterfaceNil() bool {
	return e == nil
}
