package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// EpochEconomicsStub -
type EpochEconomicsStub struct {
	ComputeEndOfEpochEconomicsCalled   func(metaBlock data.MetaHeaderHandler) (*block.Economics, error)
	ComputeEndOfEpochEconomicsV3Called func(
		metaBlock data.MetaHeaderHandler,
		prevBlockExecutionResults data.BaseMetaExecutionResultHandler,
		epochStartHandler data.EpochStartHandler,
	) (*block.Economics, error)
	VerifyRewardsPerBlockCalled func(
		metaBlock data.MetaHeaderHandler, correctedProtocolSustainability *big.Int, computedEconomics data.EconomicsHandler,
	) error
}

// ComputeEndOfEpochEconomics -
func (e *EpochEconomicsStub) ComputeEndOfEpochEconomics(metaBlock data.MetaHeaderHandler) (*block.Economics, error) {
	if e.ComputeEndOfEpochEconomicsCalled != nil {
		return e.ComputeEndOfEpochEconomicsCalled(metaBlock)
	}
	return &block.Economics{
		RewardsForProtocolSustainability: big.NewInt(0),
	}, nil
}

// ComputeEndOfEpochEconomicsV3 -
func (e *EpochEconomicsStub) ComputeEndOfEpochEconomicsV3(
	metaBlock data.MetaHeaderHandler,
	prevBlockExecutionResults data.BaseMetaExecutionResultHandler,
	epochStartHandler data.EpochStartHandler,
) (*block.Economics, error) {
	if e.ComputeEndOfEpochEconomicsV3Called != nil {
		return e.ComputeEndOfEpochEconomicsV3Called(metaBlock, prevBlockExecutionResults, epochStartHandler)
	}
	return &block.Economics{}, nil
}

// VerifyRewardsPerBlock -
func (e *EpochEconomicsStub) VerifyRewardsPerBlock(
	metaBlock data.MetaHeaderHandler,
	correctedProtocolSustainability *big.Int,
	computedEconomics data.EconomicsHandler,
) error {
	if e.VerifyRewardsPerBlockCalled != nil {
		return e.VerifyRewardsPerBlockCalled(metaBlock, correctedProtocolSustainability, computedEconomics)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochEconomicsStub) IsInterfaceNil() bool {
	return e == nil
}
