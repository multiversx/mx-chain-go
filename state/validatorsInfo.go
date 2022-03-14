package state

import (
	"bytes"
	"sync"
)

type validatorsInfo struct {
	mutex   sync.Mutex
	valInfo map[uint32][]ValidatorInfoHandler
}

func NewValidatorsInfo(input map[uint32][]*ValidatorInfo) *validatorsInfo {
	ret := &validatorsInfo{}
	ret.valInfo = make(map[uint32][]ValidatorInfoHandler, len(input))

	for shardID, valInShard := range input {
		for _, val := range valInShard {
			ret.valInfo[shardID] = append(ret.valInfo[shardID], val)
		}
	}

	return ret
}

func (vi *validatorsInfo) GetAllValidatorsInfo() []ValidatorInfoHandler {
	ret := make([]ValidatorInfoHandler, 0)

	for _, valInShard := range vi.valInfo {
		for _, val := range valInShard {
			ret = append(ret, val)
		}
	}
	return ret
}

func (vi *validatorsInfo) GetValidatorsInfoInShard(shardID uint32) []ValidatorInfoHandler {
	validatorsInShard := vi.valInfo[shardID]

	return validatorsInShard
}

func (vi *validatorsInfo) GetShardValidatorsInfo() map[uint32][]ValidatorInfoHandler {
	ret := make(map[uint32][]ValidatorInfoHandler, 0)

	for shardID, valInShard := range vi.valInfo {
		for _, val := range valInShard {
			ret[shardID] = append(ret[shardID], val)
		}
	}
	return ret
}

func (vi *validatorsInfo) Add(validatorInfo ValidatorInfoHandler) {
	//todo : handle if shard does not exist
	vi.valInfo[validatorInfo.GetShardId()] = append(vi.valInfo[validatorInfo.GetShardId()], validatorInfo)
}

func (vi *validatorsInfo) Replace(old ValidatorInfoHandler, new ValidatorInfoHandler) {
	for idx, validatorInfo := range vi.GetValidatorsInfoInShard(old.GetShardId()) {
		if bytes.Equal(validatorInfo.GetPublicKey(), old.GetPublicKey()) {
			//validatorsInfoMap.SetValidator(newValidator)
			vi.valInfo[old.GetShardId()][idx] = new
			//validatorsInfoMap[jailedValidator.ShardId][index] = newValidator
			break
		}
	}
}

func (vi *validatorsInfo) SetValidatorsInShard(shardID uint32, validators []ValidatorInfoHandler) {
	vi.valInfo[shardID] = validators
}

func (vi *validatorsInfo) Delete(shardID uint32, pubKey []byte) {
	for index, validatorInfo := range vi.GetValidatorsInfoInShard(shardID) {
		if bytes.Equal(validatorInfo.GetPublicKey(), pubKey) {
			length := len(vi.GetValidatorsInfoInShard(shardID))
			vi.valInfo[shardID][index] = vi.valInfo[shardID][length-1]
			vi.valInfo[shardID][length-1] = nil
			vi.valInfo[shardID] = vi.valInfo[shardID][:length-1]
			break
		}
	}
}

func (vi *validatorsInfo) GetMapPointer() map[uint32][]*ValidatorInfo {
	ret := make(map[uint32][]*ValidatorInfo, 0)

	for shardID, valInShard := range vi.valInfo {
		for _, val := range valInShard {
			ret[shardID] = append(ret[shardID], &ValidatorInfo{
				PublicKey:                       val.GetPublicKey(),
				ShardId:                         val.GetShardId(),
				List:                            val.GetList(),
				Index:                           val.GetIndex(),
				TempRating:                      val.GetTempRating(),
				Rating:                          val.GetRating(),
				RatingModifier:                  val.GetRatingModifier(),
				RewardAddress:                   val.GetRewardAddress(),
				LeaderSuccess:                   val.GetLeaderSuccess(),
				LeaderFailure:                   val.GetLeaderFailure(),
				ValidatorSuccess:                val.GetValidatorSuccess(),
				ValidatorFailure:                val.GetValidatorFailure(),
				ValidatorIgnoredSignatures:      val.GetValidatorIgnoredSignatures(),
				NumSelectedInSuccessBlocks:      val.GetNumSelectedInSuccessBlocks(),
				AccumulatedFees:                 val.GetAccumulatedFees(),
				TotalLeaderSuccess:              val.GetTotalLeaderSuccess(),
				TotalLeaderFailure:              val.GetTotalLeaderFailure(),
				TotalValidatorSuccess:           val.GetValidatorSuccess(),
				TotalValidatorFailure:           val.GetValidatorFailure(),
				TotalValidatorIgnoredSignatures: val.GetValidatorIgnoredSignatures(),
			})
		}
	}
	return ret
}
