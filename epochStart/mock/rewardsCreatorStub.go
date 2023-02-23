package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

// RewardsCreatorStub -
type RewardsCreatorStub struct {
	CreateRewardsMiniBlocksCalled func(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocksCalled func(
		metaBlock data.MetaHeaderHandler, validatorsInfo map[uint32][]*state.ValidatorInfo, computedEconomics *block.Economics,
	) error
	GetProtocolSustainabilityRewardsCalled func() *big.Int
	GetLocalTxCacheCalled                  func() epochStart.TransactionCacher
	CreateMarshalledDataCalled             func(body *block.Body) map[string][][]byte
	GetRewardsTxsCalled                    func(body *block.Body) map[string]data.TransactionHandler
	SaveBlockDataToStorageCalled           func(metaBlock data.MetaHeaderHandler, body *block.Body)
	DeleteBlockDataFromStorageCalled       func(metaBlock data.MetaHeaderHandler, body *block.Body)
	RemoveBlockDataFromPoolsCalled         func(metaBlock data.MetaHeaderHandler, body *block.Body)
}

// CreateRewardsMiniBlocks -
func (rcs *RewardsCreatorStub) CreateRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) (block.MiniBlockSlice, error) {
	if rcs.CreateRewardsMiniBlocksCalled != nil {
		return rcs.CreateRewardsMiniBlocksCalled(metaBlock, validatorsInfo, computedEconomics)
	}

	return nil, nil
}

// VerifyRewardsMiniBlocks -
func (rcs *RewardsCreatorStub) VerifyRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) error {
	if rcs.VerifyRewardsMiniBlocksCalled != nil {
		return rcs.VerifyRewardsMiniBlocksCalled(metaBlock, validatorsInfo, computedEconomics)
	}
	return nil
}

// GetProtocolSustainabilityRewards -
func (rcs *RewardsCreatorStub) GetProtocolSustainabilityRewards() *big.Int {
	if rcs.GetProtocolSustainabilityRewardsCalled != nil {
		return rcs.GetProtocolSustainabilityRewardsCalled()
	}

	return big.NewInt(0)
}

// GetLocalTxCache -
func (rcs *RewardsCreatorStub) GetLocalTxCache() epochStart.TransactionCacher {
	if rcs.GetLocalTxCacheCalled != nil {
		return rcs.GetLocalTxCacheCalled()
	}
	return nil
}

// CreateMarshalledData -
func (rcs *RewardsCreatorStub) CreateMarshalledData(body *block.Body) map[string][][]byte {
	if rcs.CreateMarshalledDataCalled != nil {
		return rcs.CreateMarshalledDataCalled(body)
	}
	return nil
}

// GetRewardsTxs -
func (rcs *RewardsCreatorStub) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	if rcs.GetRewardsTxsCalled != nil {
		return rcs.GetRewardsTxsCalled(body)
	}
	return nil
}

// SaveBlockDataToStorage -
func (rcs *RewardsCreatorStub) SaveBlockDataToStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if rcs.SaveBlockDataToStorageCalled != nil {
		rcs.SaveBlockDataToStorageCalled(metaBlock, body)
	}
}

// DeleteBlockDataFromStorage -
func (rcs *RewardsCreatorStub) DeleteBlockDataFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if rcs.DeleteBlockDataFromStorageCalled != nil {
		rcs.DeleteBlockDataFromStorageCalled(metaBlock, body)
	}
}

// RemoveBlockDataFromPools -
func (rcs *RewardsCreatorStub) RemoveBlockDataFromPools(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if rcs.RemoveBlockDataFromPoolsCalled != nil {
		rcs.RemoveBlockDataFromPoolsCalled(metaBlock, body)
	}
}

// IsInterfaceNil -
func (rcs *RewardsCreatorStub) IsInterfaceNil() bool {
	return rcs == nil
}
