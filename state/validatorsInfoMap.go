package state

import (
	"bytes"
	"sync"
)

type shardValidatorsInfoMap struct {
	mutex   sync.RWMutex
	valInfo map[uint32][]ValidatorInfoHandler
}

func NewShardValidatorsInfoMap(numOfShards uint32) *shardValidatorsInfoMap {
	return &shardValidatorsInfoMap{
		mutex:   sync.RWMutex{},
		valInfo: make(map[uint32][]ValidatorInfoHandler, numOfShards),
	}
}

func NewValidatorsInfo(input map[uint32][]*ValidatorInfo) *shardValidatorsInfoMap {
	ret := &shardValidatorsInfoMap{}
	ret.valInfo = make(map[uint32][]ValidatorInfoHandler, len(input))

	for shardID, valInShard := range input {
		for _, val := range valInShard {
			ret.valInfo[shardID] = append(ret.valInfo[shardID], val)
		}
	}

	return ret
}

func (vi *shardValidatorsInfoMap) GetAllValidatorsInfo() []ValidatorInfoHandler {
	ret := make([]ValidatorInfoHandler, 0)

	vi.mutex.RLock()
	validatorsMapCopy := vi.valInfo
	vi.mutex.RUnlock()

	for _, valInShard := range validatorsMapCopy {
		validatorsCopy := make([]ValidatorInfoHandler, len(valInShard))
		copy(validatorsCopy, valInShard)
		ret = append(ret, validatorsCopy...)
	}

	return ret
}

func (vi *shardValidatorsInfoMap) GetShardValidatorsInfoMap() map[uint32][]ValidatorInfoHandler {
	ret := make(map[uint32][]ValidatorInfoHandler, 0)

	vi.mutex.RLock()
	validatorsMapCopy := vi.valInfo
	vi.mutex.RUnlock()

	for shardID, valInShard := range validatorsMapCopy {
		validatorsCopy := make([]ValidatorInfoHandler, len(valInShard))
		copy(validatorsCopy, valInShard)
		ret[shardID] = validatorsCopy
	}

	return ret
}

func (vi *shardValidatorsInfoMap) Add(validator ValidatorInfoHandler) {
	shardID := validator.GetShardId()

	vi.mutex.Lock()
	vi.valInfo[shardID] = append(vi.valInfo[shardID], validator)
	vi.mutex.Unlock()
}

func (vi *shardValidatorsInfoMap) Replace(old ValidatorInfoHandler, new ValidatorInfoHandler) {
	if old.GetShardId() != new.GetShardId() {
		return
	}

	shardID := old.GetShardId()
	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	for idx, validator := range vi.valInfo[shardID] {
		if bytes.Equal(validator.GetPublicKey(), old.GetPublicKey()) {
			vi.valInfo[shardID][idx] = new
			break
		}
	}
}

func (vi *shardValidatorsInfoMap) SetValidatorsInShard(shardID uint32, validators []ValidatorInfoHandler) {
	sameShardValidators := make([]ValidatorInfoHandler, 0, len(validators))
	for _, validator := range validators {
		if validator.GetShardId() == shardID {
			sameShardValidators = append(sameShardValidators, validator)
		}
	}

	vi.mutex.Lock()
	vi.valInfo[shardID] = sameShardValidators
	vi.mutex.Unlock()
}

func (vi *shardValidatorsInfoMap) Delete(validator ValidatorInfoHandler) {
	vi.mutex.Lock()
	defer vi.mutex.Unlock()
	shardID := validator.GetShardId()

	for index, validatorInfo := range vi.valInfo[shardID] {
		if bytes.Equal(validatorInfo.GetPublicKey(), validator.GetPublicKey()) {
			length := len(vi.valInfo[shardID])
			vi.valInfo[shardID][index] = vi.valInfo[shardID][length-1]
			vi.valInfo[shardID][length-1] = nil
			vi.valInfo[shardID] = vi.valInfo[shardID][:length-1]
			break
		}
	}
}

func (vi *shardValidatorsInfoMap) GetMapPointer() map[uint32][]*ValidatorInfo {
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
