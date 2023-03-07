package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

// EpochRewardsCreatorStub -
type EpochRewardsCreatorStub struct {
	CreateRewardsMiniBlocksCalled func(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocksCalled func(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) error
	CreateMarshalledDataCalled       func(body *block.Body) map[string][][]byte
	SaveBlockDataToStorageCalled     func(metaBlock data.MetaHeaderHandler, body *block.Body)
	DeleteBlockDataFromStorageCalled func(metaBlock data.MetaHeaderHandler, body *block.Body)
	RemoveBlockDataFromPoolsCalled   func(metaBlock data.MetaHeaderHandler, body *block.Body)
	GetRewardsTxsCalled              func(body *block.Body) map[string]data.TransactionHandler
	GetProtocolSustainCalled         func() *big.Int
	GetLocalTxCacheCalled            func() epochStart.TransactionCacher
}

// GetProtocolSustainabilityRewards -
func (e *EpochRewardsCreatorStub) GetProtocolSustainabilityRewards() *big.Int {
	if e.GetProtocolSustainCalled != nil {
		return e.GetProtocolSustainCalled()
	}
	return big.NewInt(0)
}

// GetLocalTxCache -
func (e *EpochRewardsCreatorStub) GetLocalTxCache() epochStart.TransactionCacher {
	if e.GetLocalTxCacheCalled != nil {
		return e.GetLocalTxCacheCalled()
	}
	return &TxForCurrentBlockStub{}
}

// CreateRewardsMiniBlocks -
func (e *EpochRewardsCreatorStub) CreateRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) (block.MiniBlockSlice, error) {
	if e.CreateRewardsMiniBlocksCalled != nil {
		return e.CreateRewardsMiniBlocksCalled(metaBlock, validatorsInfo, computedEconomics)
	}
	return nil, nil
}

// VerifyRewardsMiniBlocks -
func (e *EpochRewardsCreatorStub) VerifyRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) error {
	if e.VerifyRewardsMiniBlocksCalled != nil {
		return e.VerifyRewardsMiniBlocksCalled(metaBlock, validatorsInfo, computedEconomics)
	}
	return nil
}

// CreateMarshalledData -
func (e *EpochRewardsCreatorStub) CreateMarshalledData(body *block.Body) map[string][][]byte {
	if e.CreateMarshalledDataCalled != nil {
		return e.CreateMarshalledDataCalled(body)
	}
	return nil
}

// GetRewardsTxs -
func (e *EpochRewardsCreatorStub) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	if e.GetRewardsTxsCalled != nil {
		return e.GetRewardsTxsCalled(body)
	}
	return nil
}

// SaveBlockDataToStorage -
func (e *EpochRewardsCreatorStub) SaveBlockDataToStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if e.SaveBlockDataToStorageCalled != nil {
		e.SaveBlockDataToStorageCalled(metaBlock, body)
	}
}

// DeleteBlockDataFromStorage -
func (e *EpochRewardsCreatorStub) DeleteBlockDataFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if e.DeleteBlockDataFromStorageCalled != nil {
		e.DeleteBlockDataFromStorageCalled(metaBlock, body)
	}
}

// IsInterfaceNil -
func (e *EpochRewardsCreatorStub) IsInterfaceNil() bool {
	return e == nil
}

// RemoveBlockDataFromPools -
func (e *EpochRewardsCreatorStub) RemoveBlockDataFromPools(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if e.RemoveBlockDataFromPoolsCalled != nil {
		e.RemoveBlockDataFromPoolsCalled(metaBlock, body)
	}
}
