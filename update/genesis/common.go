package genesis

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
)

// TODO: create a structure or use this function also in process/peer/process.go
func getValidatorDataFromLeaves(
	leavesChannels *common.TrieIteratorChannels,
	marshalizer marshal.Marshalizer,
) (state.ShardValidatorsInfoMapHandler, error) {
	validators := state.NewShardValidatorsInfoMap()
	for pa := range leavesChannels.LeavesChan {
		peerAccount, err := unmarshalPeer(pa, marshalizer)
		if err != nil {
			return nil, err
		}

		validatorInfoData := peerAccountToValidatorInfo(peerAccount)
		err = validators.Add(validatorInfoData)
		if err != nil {
			return nil, err
		}
	}

	err := leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func unmarshalPeer(peerAccountData core.KeyValueHolder, marshalizer marshal.Marshalizer) (state.PeerAccountHandler, error) {
	peerAccount, err := accounts.NewPeerAccount(peerAccountData.Key())
	if err != nil {
		return nil, err
	}
	err = marshalizer.Unmarshal(peerAccount, peerAccountData.Value())
	if err != nil {
		return nil, err
	}
	return peerAccount, nil
}

func peerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                  peerAccount.AddressBytes(),
		ShardId:                    peerAccount.GetShardId(),
		List:                       getActualList(peerAccount),
		PreviousList:               peerAccount.GetPreviousList(),
		Index:                      peerAccount.GetIndexInList(),
		PreviousIndex:              peerAccount.GetPreviousIndexInList(),
		TempRating:                 peerAccount.GetTempRating(),
		Rating:                     peerAccount.GetRating(),
		RewardAddress:              peerAccount.GetRewardAddress(),
		LeaderSuccess:              peerAccount.GetLeaderSuccessRate().GetNumSuccess(),
		LeaderFailure:              peerAccount.GetLeaderSuccessRate().GetNumFailure(),
		ValidatorSuccess:           peerAccount.GetValidatorSuccessRate().GetNumSuccess(),
		ValidatorFailure:           peerAccount.GetValidatorSuccessRate().GetNumFailure(),
		TotalLeaderSuccess:         peerAccount.GetTotalLeaderSuccessRate().GetNumSuccess(),
		TotalLeaderFailure:         peerAccount.GetTotalLeaderSuccessRate().GetNumFailure(),
		TotalValidatorSuccess:      peerAccount.GetTotalValidatorSuccessRate().GetNumSuccess(),
		TotalValidatorFailure:      peerAccount.GetTotalValidatorSuccessRate().GetNumFailure(),
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
