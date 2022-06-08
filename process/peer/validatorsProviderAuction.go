package peer

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
)

// GetAuctionList returns an array containing the validators that are currently in the auction list
func (vp *validatorsProvider) GetAuctionList() []*common.AuctionListValidatorAPIResponse {
	validatorsMap, _ := vp.getValidatorsInfo() //todo: error
	defer vp.stakingDataProvider.Clean()

	for _, validator := range validatorsMap.GetAllValidatorsInfo() {
		_ = vp.stakingDataProvider.FillValidatorInfo(validator) // todo: error
	}

	vp.auctionLock.RLock()
	randomness := vp.cachedRandomness
	vp.auctionLock.RUnlock()
	_ = vp.auctionListSelector.SelectNodesFromAuctionList(validatorsMap, randomness) //todo : error + randomness

	auctionListValidators := make([]*common.AuctionListValidatorAPIResponse, 0)

	for ownerPubKey, ownerData := range vp.stakingDataProvider.GetOwnersData() {
		if ownerData.Qualified && ownerData.NumAuctionNodes > 0 {
			auctionListValidators = append(auctionListValidators, &common.AuctionListValidatorAPIResponse{
				Owner: vp.addressPubKeyConverter.Encode([]byte(ownerPubKey)),
				// todo: if his node from auction is selected, add necessary data
			})
		}
	}

	return auctionListValidators
}

func (vp *validatorsProvider) getValidatorsInfo() (state.ShardValidatorsInfoMapHandler, error) {
	vp.auctionLock.RLock()
	shouldUpdate := time.Since(vp.lastValidatorsInfoCacheUpdate) > vp.cacheRefreshIntervalDuration
	vp.auctionLock.RUnlock()

	if shouldUpdate {
		err := vp.updateValidatorsInfoCache()
		if err != nil {
			return nil, err
		}
	}

	vp.auctionLock.RLock()
	defer vp.auctionLock.RUnlock()

	return cloneValidatorsMap(vp.cachedValidatorsMap)
}

func (vp *validatorsProvider) updateValidatorsInfoCache() error {
	rootHash, err := vp.validatorStatistics.RootHash()
	if err != nil {
		return err
	}

	validatorsMap, err := vp.validatorStatistics.GetValidatorInfoForRootHash(rootHash)
	if err != nil {
		return err
	}

	vp.auctionLock.Lock()
	defer vp.auctionLock.Unlock()

	vp.lastValidatorsInfoCacheUpdate = time.Now()
	vp.cachedValidatorsMap, err = cloneValidatorsMap(validatorsMap)
	vp.cachedRandomness = rootHash
	if err != nil {
		return err
	}

	return nil
}

func cloneValidatorsMap(validatorsMap state.ShardValidatorsInfoMapHandler) (state.ShardValidatorsInfoMapHandler, error) {
	ret := state.NewShardValidatorsInfoMap()
	for _, validator := range validatorsMap.GetAllValidatorsInfo() {
		err := ret.Add(validator.ShallowClone())
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}
