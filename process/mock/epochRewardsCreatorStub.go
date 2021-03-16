package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochRewardsCreatorStub -
type EpochRewardsCreatorStub struct {
	CreateRewardsMiniBlocksCalled func(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocksCalled func(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) error
	CreateMarshalizedDataCalled    func(body *block.Body) map[string][][]byte
	SaveTxBlockToStorageCalled     func(metaBlock data.MetaHeaderHandler, body *block.Body)
	DeleteTxsFromStorageCalled     func(metaBlock data.MetaHeaderHandler, body *block.Body)
	RemoveBlockDataFromPoolsCalled func(metaBlock data.MetaHeaderHandler, body *block.Body)
	GetRewardsTxsCalled            func(body *block.Body) map[string]data.TransactionHandler
	GetProtocolSustainCalled       func() *big.Int
	GetLocalTxCacheCalled          func() epochStart.TransactionCacher
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

// CreateMarshalizedData -
func (e *EpochRewardsCreatorStub) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	if e.CreateMarshalizedDataCalled != nil {
		return e.CreateMarshalizedDataCalled(body)
	}
	return nil
}

// GetRewardsTxs --
func (e *EpochRewardsCreatorStub) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	if e.GetRewardsTxsCalled != nil {
		return e.GetRewardsTxsCalled(body)
	}
	return nil
}

// SaveTxBlockToStorage -
func (e *EpochRewardsCreatorStub) SaveTxBlockToStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if e.SaveTxBlockToStorageCalled != nil {
		e.SaveTxBlockToStorageCalled(metaBlock, body)
	}
}

// DeleteTxsFromStorage -
func (e *EpochRewardsCreatorStub) DeleteTxsFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if e.DeleteTxsFromStorageCalled != nil {
		e.DeleteTxsFromStorageCalled(metaBlock, body)
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
