package genesis

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"

	"github.com/multiversx/mx-chain-go/state"
)

// TODO: create a structure or use this function also in process/peer/process.go
func getValidatorDataFromLeaves(
	leavesChannels *common.TrieIteratorChannels,
	marshalizer marshal.Marshalizer,
) (state.ShardValidatorsInfoMapHandler, error) {
	validators := state.NewShardValidatorsInfoMap()
	for pa := range leavesChannels.LeavesChan {
		peerAccount, err := unmarshalPeer(pa.Value(), marshalizer)
		if err != nil {
			return nil, err
		}

		validatorInfoData := peerAccountToValidatorInfo(peerAccount)
		err = validators.Add(validatorInfoData)
		if err != nil {
			return nil, err
		}
	}

	err := common.GetErrorFromChanNonBlocking(leavesChannels.ErrChan)
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func unmarshalPeer(pa []byte, marshalizer marshal.Marshalizer) (state.PeerAccountHandler, error) {
	peerAccount := state.NewEmptyPeerAccount()
	err := marshalizer.Unmarshal(peerAccount, pa)
	if err != nil {
		return nil, err
	}
	return peerAccount, nil
}

func peerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                  peerAccount.GetBLSPublicKey(),
		ShardId:                    peerAccount.GetShardId(),
		List:                       getActualList(peerAccount),
		PreviousList:               peerAccount.GetPreviousList(),
		Index:                      peerAccount.GetIndexInList(),
		PreviousIndex:              peerAccount.GetPreviousIndexInList(),
		TempRating:                 peerAccount.GetTempRating(),
		Rating:                     peerAccount.GetRating(),
		RewardAddress:              peerAccount.GetRewardAddress(),
		LeaderSuccess:              peerAccount.GetLeaderSuccessRate().NumSuccess,
		LeaderFailure:              peerAccount.GetLeaderSuccessRate().NumFailure,
		ValidatorSuccess:           peerAccount.GetValidatorSuccessRate().NumSuccess,
		ValidatorFailure:           peerAccount.GetValidatorSuccessRate().NumFailure,
		TotalLeaderSuccess:         peerAccount.GetTotalLeaderSuccessRate().NumSuccess,
		TotalLeaderFailure:         peerAccount.GetTotalLeaderSuccessRate().NumFailure,
		TotalValidatorSuccess:      peerAccount.GetTotalValidatorSuccessRate().NumSuccess,
		TotalValidatorFailure:      peerAccount.GetTotalValidatorSuccessRate().NumFailure,
		NumSelectedInSuccessBlocks: peerAccount.GetNumSelectedInSuccessBlocks(),
		AccumulatedFees:            big.NewInt(0).Set(peerAccount.GetAccumulatedFees()),
	}
}

func getActualList(peerAccount state.PeerAccountHandler) string {
	savedList := peerAccount.GetList()
	if peerAccount.GetUnStakedEpoch() == common.DefaultUnstakedEpoch {
		if savedList == string(common.InactiveList) {
			return string(common.JailedList)
		}
		return savedList
	}
	if savedList == string(common.InactiveList) {
		return savedList
	}

	return string(common.LeavingList)
}

func shouldExportValidator(validator state.ValidatorInfoHandler, allowedLists []common.PeerType) bool {
	validatorList := validator.GetList()

	for _, list := range allowedLists {
		if validatorList == string(list) {
			return true
		}
	}

	return false
}
