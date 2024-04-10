package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type shardValidatorsInfoMap struct {
	mutex      sync.RWMutex
	valInfoMap map[uint32][]ValidatorInfoHandler
}

// NewShardValidatorsInfoMap creates an instance of shardValidatorsInfoMap which manages a
// <shardID, validatorsInfo> map internally
func NewShardValidatorsInfoMap() *shardValidatorsInfoMap {
	return &shardValidatorsInfoMap{
		mutex:      sync.RWMutex{},
		valInfoMap: make(map[uint32][]ValidatorInfoHandler),
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
			return validator.ShallowClone()
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
	log.Debug("shardValidatorsInfoMap.Replace",
		"old validator", hex.EncodeToString(old.GetPublicKey()), "shard", old.GetShardId(), "list", old.GetList(),
		"with new validator", hex.EncodeToString(new.GetPublicKey()), "shard", new.GetShardId(), "list", new.GetList(),
	)

	replaced := vi.ReplaceValidatorByKey(old.GetPublicKey(), new, shardID)
	if replaced {
		return nil
	}

	return fmt.Errorf("old %w: %s when trying to replace it with %s",
		ErrValidatorNotFound,
		hex.EncodeToString(old.GetPublicKey()),
		hex.EncodeToString(new.GetPublicKey()),
	)
}

// ReplaceValidatorByKey will replace an existing ValidatorInfoHandler with a new one, based on the provided blsKey for the old record.
func (vi *shardValidatorsInfoMap) ReplaceValidatorByKey(oldBlsKey []byte, new ValidatorInfoHandler, shardID uint32) bool {
	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	for idx, validator := range vi.valInfoMap[shardID] {
		if bytes.Equal(validator.GetPublicKey(), oldBlsKey) {
			vi.valInfoMap[shardID][idx] = new
			return true
		}
	}
	return false
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
	vi.DeleteByKey(validator.GetPublicKey(), shardID)
	return nil
}

// DeleteByKey will delete the provided blsKey from the internally stored map, if found.
func (vi *shardValidatorsInfoMap) DeleteByKey(blsKey []byte, shardID uint32) {
	vi.mutex.Lock()
	defer vi.mutex.Unlock()

	for index, validatorInfo := range vi.valInfoMap[shardID] {
		if bytes.Equal(validatorInfo.GetPublicKey(), blsKey) {
			length := len(vi.valInfoMap[shardID])
			vi.valInfoMap[shardID][index] = vi.valInfoMap[shardID][length-1]
			vi.valInfoMap[shardID][length-1] = nil
			vi.valInfoMap[shardID] = vi.valInfoMap[shardID][:length-1]
			break
		}
	}
}
