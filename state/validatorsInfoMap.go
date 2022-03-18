package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
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
func Replace(oldMap, newMap map[uint32][]*ValidatorInfo) {
	for shardID := range oldMap {
		delete(oldMap, shardID)
	}

	for shardID, validatorsInShard := range newMap {
		oldMap[shardID] = validatorsInShard
	}

}

// GetAllValidatorsInfo returns a []ValidatorInfoHandler copy with validators from all shards.
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

// GetShardValidatorsInfoMap returns a <shard, ValidatorInfoHandler> map copy of internally stored data
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

// Add adds a ValidatorInfoHandler in its corresponding shardID
func (vi *shardValidatorsInfoMap) Add(validator ValidatorInfoHandler) error {
	if check.IfNil(validator) {
		return ErrNilValidatorInfo
	}

	shardID := validator.GetShardId()
	vi.mutex.Lock()
	vi.valInfoMap[shardID] = append(vi.valInfoMap[shardID], validator)
	vi.mutex.Unlock()

	return nil
}

// GetValidator returns a ValidatorInfoHandler copy with the provided blsKey,
// if it is present in the map, otherwise returns nil
func (vi *shardValidatorsInfoMap) GetValidator(blsKey []byte) ValidatorInfoHandler {
	for _, validator := range vi.GetAllValidatorsInfo() {
		if bytes.Equal(validator.GetPublicKey(), blsKey) {
			return validator
		}
	}

	return nil
}

// Replace will replace an existing ValidatorInfoHandler with a new one. The old and new validator
// shall be in the same shard. If the old validator is not found in the map, an error is returned
func (vi *shardValidatorsInfoMap) Replace(old ValidatorInfoHandler, new ValidatorInfoHandler) error {
	if check.IfNil(old) {
		return fmt.Errorf("%w for old validator in shardValidatorsInfoMap.Replace", ErrNilValidatorInfo)
	}
	if check.IfNil(new) {
		return fmt.Errorf("%w for new validator in shardValidatorsInfoMap.Replace", ErrNilValidatorInfo)
	}
	if old.GetShardId() != new.GetShardId() {
		return fmt.Errorf("%w when trying to replace %s from shard %v with %s from shard %v",
			ErrValidatorsDifferentShards,
			hex.EncodeToString(old.GetPublicKey()),
			old.GetShardId(),
			hex.EncodeToString(new.GetPublicKey()),
			new.GetShardId(),
		)
	}

	shardID := old.GetShardId()

	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	for idx, validator := range vi.valInfoMap[shardID] {
		if bytes.Equal(validator.GetPublicKey(), old.GetPublicKey()) {
			vi.valInfoMap[shardID][idx] = new
			return nil
		}
	}

	return fmt.Errorf("old %w: %s when trying to replace it with %s",
		ErrValidatorNotFound,
		hex.EncodeToString(old.GetPublicKey()),
		hex.EncodeToString(new.GetPublicKey()),
	)
}

// SetValidatorsInShard resets all validators saved in a specific shard with the provided []ValidatorInfoHandler.
// Before setting them, it checks that provided validators have the same shardID as the one provided.
func (vi *shardValidatorsInfoMap) SetValidatorsInShard(shardID uint32, validators []ValidatorInfoHandler) error {
	sameShardValidators := make([]ValidatorInfoHandler, 0, len(validators))
	for idx, validator := range validators {
		if check.IfNil(validator) {
			return fmt.Errorf("%w in shardValidatorsInfoMap.SetValidatorsInShard at index %d",
				ErrNilValidatorInfo,
				idx,
			)
		}
		if validator.GetShardId() != shardID {
			return fmt.Errorf("%w, %s is in shard %d, but should be set in shard %d in shardValidatorsInfoMap.SetValidatorsInShard",
				ErrValidatorsDifferentShards,
				hex.EncodeToString(validator.GetPublicKey()),
				validator.GetShardId(),
				shardID,
			)
		}
		sameShardValidators = append(sameShardValidators, validator)
	}

	vi.mutex.Lock()
	vi.valInfoMap[shardID] = sameShardValidators
	vi.mutex.Unlock()

	return nil
}

// Delete will delete the provided validator from the internally stored map, if found.
// The validators slice at the corresponding shardID key will be re-sliced, without reordering
func (vi *shardValidatorsInfoMap) Delete(validator ValidatorInfoHandler) error {
	if check.IfNil(validator) {
		return ErrNilValidatorInfo
	}

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

	return nil
}

// TODO: Delete this once map[uint32][]*ValidatorInfo is completely replaced with new interface

// GetValInfoPointerMap returns a <shardID, []*ValidatorInfo> from internally stored data
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
