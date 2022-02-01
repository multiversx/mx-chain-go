package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/state"
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
	CreateMarshalizedDataCalled            func(body *block.Body) map[string][][]byte
	GetRewardsTxsCalled                    func(body *block.Body) map[string]data.TransactionHandler
	SaveTxBlockToStorageCalled             func(metaBlock data.MetaHeaderHandler, body *block.Body)
	DeleteTxsFromStorageCalled             func(metaBlock data.MetaHeaderHandler, body *block.Body)
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

// CreateMarshalizedData -
func (rcs *RewardsCreatorStub) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	if rcs.CreateMarshalizedDataCalled != nil {
		return rcs.CreateMarshalizedDataCalled(body)
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

// SaveTxBlockToStorage -
func (rcs *RewardsCreatorStub) SaveTxBlockToStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if rcs.SaveTxBlockToStorageCalled != nil {
		rcs.SaveTxBlockToStorageCalled(metaBlock, body)
	}
}

// DeleteTxsFromStorage -
func (rcs *RewardsCreatorStub) DeleteTxsFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	if rcs.DeleteTxsFromStorageCalled != nil {
		rcs.DeleteTxsFromStorageCalled(metaBlock, body)
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
