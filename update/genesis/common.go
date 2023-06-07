package genesis

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
)

// TODO: create a structure or use this function also in process/peer/process.go
func getValidatorDataFromLeaves(
	leavesChannels *common.TrieIteratorChannels,
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
) (map[uint32][]*state.ValidatorInfo, error) {

	validators := make(map[uint32][]*state.ValidatorInfo, shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		validators[i] = make([]*state.ValidatorInfo, 0)
	}
	validators[core.MetachainShardId] = make([]*state.ValidatorInfo, 0)

	for pa := range leavesChannels.LeavesChan {
		peerAccount, err := unmarshalPeer(pa.Value(), marshalizer)
		if err != nil {
			return nil, err
		}

		currentShardId := peerAccount.GetShardId()
		validatorInfoData := peerAccountToValidatorInfo(peerAccount)
		validators[currentShardId] = append(validators[currentShardId], validatorInfoData)
	}

	err := leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func unmarshalPeer(pa []byte, marshalizer marshal.Marshalizer) (common.PeerAccountHandler, error) {
	peerAccount := accounts.NewEmptyPeerAccount()
	err := marshalizer.Unmarshal(peerAccount, pa)
	if err != nil {
		return nil, err
	}
	return peerAccount, nil
}

func peerAccountToValidatorInfo(peerAccount common.PeerAccountHandler) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                  peerAccount.AddressBytes(),
		ShardId:                    peerAccount.GetShardId(),
		List:                       getActualList(peerAccount),
		Index:                      peerAccount.GetIndexInList(),
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

func getActualList(peerAccount common.PeerAccountHandler) string {
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

func shouldExportValidator(validator *state.ValidatorInfo, allowedLists []common.PeerType) bool {
	validatorList := validator.GetList()

	for _, list := range allowedLists {
		if validatorList == string(list) {
			return true
		}
	}

	return false
}
