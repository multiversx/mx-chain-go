package peer

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/state"
)

// GetAuctionList returns an array containing the validators that are currently in the auction list
func (vp *validatorsProvider) GetAuctionList() ([]*common.AuctionListValidatorAPIResponse, error) {
	err := vp.updateAuctionListCacheIfNeeded()
	if err != nil {
		return nil, err
	}

	vp.auctionMutex.RLock()
	ret := make([]*common.AuctionListValidatorAPIResponse, len(vp.cachedAuctionValidators))
	copy(ret, vp.cachedAuctionValidators)
	vp.auctionMutex.RUnlock()

	return ret, nil
}

func (vp *validatorsProvider) updateAuctionListCacheIfNeeded() error {
	vp.auctionMutex.RLock()
	shouldUpdate := time.Since(vp.lastAuctionCacheUpdate) > vp.cacheRefreshIntervalDuration
	vp.auctionMutex.RUnlock()

	if shouldUpdate {
		return vp.updateAuctionListCache()
	}

	return nil
}

func (vp *validatorsProvider) updateAuctionListCache() error {
	rootHash := vp.validatorStatistics.LastFinalizedRootHash()
	if len(rootHash) == 0 {
		log.Warn("Finalize roothash is 0")
		return nil
	}

	validatorsMap, err := vp.validatorStatistics.GetValidatorInfoForRootHash(rootHash)
	if err != nil {
		validatorsMap = state.NewShardValidatorsInfoMap()
		log.Error("Error getting validators for roothash", "error", err, "roothash", hex.EncodeToString(rootHash))
	}

	vp.auctionMutex.Lock()
	vp.cachedRandomness = rootHash
	vp.auctionMutex.Unlock()

	newCache, err := vp.createValidatorsAuctionCache(validatorsMap)
	if err != nil {
		return err
	}

	vp.auctionMutex.Lock()
	vp.lastAuctionCacheUpdate = time.Now()
	vp.cachedAuctionValidators = newCache
	vp.auctionMutex.Unlock()

	return nil
}

func (vp *validatorsProvider) createValidatorsAuctionCache(validatorsMap state.ShardValidatorsInfoMapHandler) ([]*common.AuctionListValidatorAPIResponse, error) {
	defer vp.stakingDataProvider.Clean()

	err := vp.fillAllValidatorsInfo(validatorsMap)
	if err != nil {
		return nil, err
	}

	selectedNodes, err := vp.getSelectedNodesFromAuction(validatorsMap)
	if err != nil {
		return nil, err
	}

	auctionListValidators := vp.getAuctionListValidatorsAPIResponse(selectedNodes)
	sortList(auctionListValidators)
	return auctionListValidators, nil
}

func (vp *validatorsProvider) fillAllValidatorsInfo(validatorsMap state.ShardValidatorsInfoMapHandler) error {
	for _, validator := range validatorsMap.GetAllValidatorsInfo() {
		err := vp.stakingDataProvider.FillValidatorInfo(validator)
		if err != nil {
			return err
		}
	}

	_, _, err := vp.stakingDataProvider.ComputeUnQualifiedNodes(validatorsMap)
	return err
}

func (vp *validatorsProvider) getSelectedNodesFromAuction(validatorsMap state.ShardValidatorsInfoMapHandler) ([]state.ValidatorInfoHandler, error) {
	vp.auctionMutex.RLock()
	randomness := vp.cachedRandomness
	vp.auctionMutex.RUnlock()

	err := vp.auctionListSelector.SelectNodesFromAuctionList(validatorsMap, randomness)
	if err != nil {
		return nil, err
	}

	selectedNodes := make([]state.ValidatorInfoHandler, 0)
	for _, validator := range validatorsMap.GetAllValidatorsInfo() {
		if validator.GetList() == string(common.SelectedFromAuctionList) {
			selectedNodes = append(selectedNodes, validator.ShallowClone())
		}
	}

	return selectedNodes, nil
}

func sortList(list []*common.AuctionListValidatorAPIResponse) {
	sort.SliceStable(list, func(i, j int) bool {
		qualifiedTopUpValidator1, _ := big.NewInt(0).SetString(list[i].QualifiedTopUp, 10)
		qualifiedTopUpValidator2, _ := big.NewInt(0).SetString(list[j].QualifiedTopUp, 10)

		return qualifiedTopUpValidator1.Cmp(qualifiedTopUpValidator2) > 0
	})
}

func (vp *validatorsProvider) getAuctionListValidatorsAPIResponse(selectedNodes []state.ValidatorInfoHandler) []*common.AuctionListValidatorAPIResponse {
	auctionListValidators := make([]*common.AuctionListValidatorAPIResponse, 0)

	for ownerPubKey, ownerData := range vp.stakingDataProvider.GetOwnersData() {
		if ownerData.NumAuctionNodes > 0 {
			auctionValidator := &common.AuctionListValidatorAPIResponse{
				Owner:          vp.addressPubKeyConverter.Encode([]byte(ownerPubKey)),
				NumStakedNodes: ownerData.NumStakedNodes,
				TotalTopUp:     ownerData.TotalTopUp.String(),
				TopUpPerNode:   ownerData.TopUpPerNode.String(),
				QualifiedTopUp: ownerData.TopUpPerNode.String(),
				AuctionList:    make([]*common.AuctionNode, 0, ownerData.NumAuctionNodes),
			}

			vp.fillAuctionQualifiedValidatorAPIData(selectedNodes, ownerData, auctionValidator)
			auctionListValidators = append(auctionListValidators, auctionValidator)
		}
	}

	return auctionListValidators
}

func (vp *validatorsProvider) fillAuctionQualifiedValidatorAPIData(
	selectedNodes []state.ValidatorInfoHandler,
	ownerData *epochStart.OwnerData,
	auctionValidatorAPI *common.AuctionListValidatorAPIResponse,
) {
	auctionValidatorAPI.AuctionList = make([]*common.AuctionNode, 0, ownerData.NumAuctionNodes)
	numOwnerQualifiedNodes := int64(0)
	for _, nodeInAuction := range ownerData.AuctionList {
		auctionNode := &common.AuctionNode{
			BlsKey:    vp.addressPubKeyConverter.Encode(nodeInAuction.GetPublicKey()),
			Qualified: false,
		}
		if ownerData.Qualified && contains(selectedNodes, nodeInAuction) {
			auctionNode.Qualified = true
			numOwnerQualifiedNodes++
		}

		auctionValidatorAPI.AuctionList = append(auctionValidatorAPI.AuctionList, auctionNode)
	}

	if numOwnerQualifiedNodes > 0 {
		activeNodes := big.NewInt(ownerData.NumActiveNodes)
		qualifiedNodes := big.NewInt(numOwnerQualifiedNodes)
		ownerRemainingNodes := big.NewInt(0).Add(activeNodes, qualifiedNodes)
		auctionValidatorAPI.QualifiedTopUp = big.NewInt(0).Div(ownerData.TotalTopUp, ownerRemainingNodes).String()
	}
}

func contains(list []state.ValidatorInfoHandler, validator state.ValidatorInfoHandler) bool {
	for _, val := range list {
		if bytes.Equal(val.GetPublicKey(), validator.GetPublicKey()) {
			return true
		}
	}
	return false
}
