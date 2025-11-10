package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// EpochStartDataCreatorStub -
type EpochStartDataCreatorStub struct {
	CreateEpochStartDataCalled             func() (*block.EpochStart, error)
	VerifyEpochStartDataForMetablockCalled func(metaBlock data.MetaHeaderHandler) error
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
