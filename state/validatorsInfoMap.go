package state

import (
	"bytes"
	"sync"
)

type shardValidatorsInfoMap struct {
	mutex      sync.RWMutex
	valInfoMap map[uint32][]ValidatorInfoHandler
}

// NewShardValidatorsInfoMap creates an instance of shardValidatorsInfoMap which manages a
// <shardID, validatorsInfo> map internally
func NewShardValidatorsInfoMap(numOfShards uint32) *shardValidatorsInfoMap {
	return &shardValidatorsInfoMap{
		mutex:      sync.RWMutex{},
		valInfoMap: make(map[uint32][]ValidatorInfoHandler, numOfShards),
	}
}

// TODO: Delete these 2 functions once map[uint32][]*ValidatorInfo is completely replaced with new interface

// CreateShardValidatorsMap creates an instance of shardValidatorsInfoMap which manages a shard validator
// info map internally.
func CreateShardValidatorsMap(input map[uint32][]*ValidatorInfo) *shardValidatorsInfoMap {
	ret := &shardValidatorsInfoMap{valInfoMap: make(map[uint32][]ValidatorInfoHandler, len(input))}

	for shardID, valInShard := range input {
		for _, val := range valInShard {
			ret.valInfoMap[shardID] = append(ret.valInfoMap[shardID], val)
		}
	}

	return ret
}

// Replace will replace src with dst map
func Replace(src, dest map[uint32][]*ValidatorInfo) {
	for shardID := range src {
		delete(src, shardID)
	}

	for shardID, validatorsInShard := range src {
		dest[shardID] = validatorsInShard
	}

}

// GetAllValidatorsInfo returns a ValidatorInfoHandler copy slice with validators from all shards.
func (vi *shardValidatorsInfoMap) GetAllValidatorsInfo() []ValidatorInfoHandler {
	ret := make([]ValidatorInfoHandler, 0)

	vi.mutex.RLock()
	defer vi.mutex.RUnlock()

	for _, validatorsInShard := range vi.valInfoMap {
		validatorsCopy := make([]ValidatorInfoHandler, len(validatorsInShard))
		copy(validatorsCopy, validatorsInShard)
		ret = append(ret, validatorsCopy...)
	}

	return ret
}

// GetShardValidatorsInfoMap returns a <shard, ValidatorInfoHandler> copy map of internally stored data
func (vi *shardValidatorsInfoMap) GetShardValidatorsInfoMap() map[uint32][]ValidatorInfoHandler {
	ret := make(map[uint32][]ValidatorInfoHandler, len(vi.valInfoMap))

	vi.mutex.RLock()
	defer vi.mutex.RUnlock()

	for shardID, validatorsInShard := range vi.valInfoMap {
		validatorsCopy := make([]ValidatorInfoHandler, len(validatorsInShard))
		copy(validatorsCopy, validatorsInShard)
		ret[shardID] = validatorsCopy
	}

	return ret
}

// Add adds a new ValidatorInfoHandler in its corresponding shardID, if it doesn't already exists
func (vi *shardValidatorsInfoMap) Add(validator ValidatorInfoHandler) {
	if vi.GetValidator(validator.GetPublicKey()) != nil {
		return
	}

	shardID := validator.GetShardId()

	vi.mutex.Lock()
	vi.valInfoMap[shardID] = append(vi.valInfoMap[shardID], validator)
	vi.mutex.Unlock()
}

// GetValidator returns a ValidatorInfoHandler with the provided blsKey, if it is present in the map
func (vi *shardValidatorsInfoMap) GetValidator(blsKey []byte) ValidatorInfoHandler {
	for _, validator := range vi.GetAllValidatorsInfo() {
		if bytes.Equal(validator.GetPublicKey(), blsKey) {
			return validator
		}
	}

	return nil
}

// Replace will replace an existing ValidatorInfoHandler with a new one. The old and new validator
// shall be in the same shard and have the same public key.
func (vi *shardValidatorsInfoMap) Replace(old ValidatorInfoHandler, new ValidatorInfoHandler) {
	if old.GetShardId() != new.GetShardId() {
		return
	}

	shardID := old.GetShardId()

	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	for idx, validator := range vi.valInfoMap[shardID] {
		if bytes.Equal(validator.GetPublicKey(), old.GetPublicKey()) {
			vi.valInfoMap[shardID][idx] = new
			break
		}
	}
}

// SetValidatorsInShard resets all validators saved in a specific shard with the provided []ValidatorInfoHandler.
// Before setting them, it checks that provided validators have the same shardID as the one provided.
func (vi *shardValidatorsInfoMap) SetValidatorsInShard(shardID uint32, validators []ValidatorInfoHandler) {
	sameShardValidators := make([]ValidatorInfoHandler, 0, len(validators))
	for _, validator := range validators {
		if validator.GetShardId() == shardID {
			sameShardValidators = append(sameShardValidators, validator)
		}
	}

	vi.mutex.Lock()
	vi.valInfoMap[shardID] = sameShardValidators
	vi.mutex.Unlock()
}

// Delete will delete the provided validator from the internally stored map. The validators slice at the
// corresponding shardID key will be re-sliced, without reordering
func (vi *shardValidatorsInfoMap) Delete(validator ValidatorInfoHandler) {
	shardID := validator.GetShardId()

	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	for index, validatorInfo := range vi.valInfoMap[shardID] {
		if bytes.Equal(validatorInfo.GetPublicKey(), validator.GetPublicKey()) {
			length := len(vi.valInfoMap[shardID])
			vi.valInfoMap[shardID][index] = vi.valInfoMap[shardID][length-1]
			vi.valInfoMap[shardID][length-1] = nil
			vi.valInfoMap[shardID] = vi.valInfoMap[shardID][:length-1]
			break
		}
	}
}

// TODO: Delete this once map[uint32][]*ValidatorInfo is completely replaced with new interface

// GetValInfoPointerMap returns a <shardID, []validators> from internally stored data
func (vi *shardValidatorsInfoMap) GetValInfoPointerMap() map[uint32][]*ValidatorInfo {
	ret := make(map[uint32][]*ValidatorInfo, 0)

	for shardID, valInShard := range vi.valInfoMap {
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
