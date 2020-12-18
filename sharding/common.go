package sharding

import (
	"encoding/hex"
	"strconv"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
)

var log = logger.GetOrCreate("sharding")

func computeStartIndexAndNumAppearancesForValidator(expEligibleList []uint32, idx int64) (int64, int64) {
	val := expEligibleList[idx]
	startIdx := int64(0)
	listLen := int64(len(expEligibleList))

	for i := idx - 1; i >= 0; i-- {
		if expEligibleList[i] != val {
			startIdx = i + 1
			break
		}
	}

	endIdx := listLen - 1
	for i := idx + 1; i < listLen; i++ {
		if expEligibleList[i] != val {
			endIdx = i - 1
			break
		}
	}

	return startIdx, endIdx - startIdx + 1
}

func displayValidatorsForRandomness(validators []Validator, randomness []byte) {
	strValidators := ""

	for _, v := range validators {
		strValidators += "\n" + hex.EncodeToString(v.PubKey())
	}

	log.Trace("selectValidators", "randomness", randomness, "validators", strValidators)
}

func displayNodesConfiguration(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving map[uint32][]Validator,
	actualRemaining map[uint32][]Validator,
	nbShards uint32,
) {
	for shard := uint32(0); shard <= nbShards; shard++ {
		shardID := shard
		if shardID == nbShards {
			shardID = core.MetachainShardId
		}
		for _, v := range eligible[shardID] {
			pk := v.PubKey()
			log.Debug("eligible", "pk", pk, "shardID", shardID)
		}
		for _, v := range waiting[shardID] {
			pk := v.PubKey()
			log.Debug("waiting", "pk", pk, "shardID", shardID)
		}
		for _, v := range leaving[shardID] {
			pk := v.PubKey()
			log.Debug("leaving", "pk", pk, "shardID", shardID)
		}
		for _, v := range actualRemaining[shardID] {
			pk := v.PubKey()
			log.Debug("actually remaining", "pk", pk, "shardID", shardID)
		}
	}
}

// SerializableValidatorsToValidators creates the validator map from serializable validator map
func SerializableValidatorsToValidators(nodeRegistryValidators map[string][]*SerializableValidator) (map[uint32][]Validator, error) {
	validators := make(map[uint32][]Validator)
	for shardId, shardValidators := range nodeRegistryValidators {
		newValidators, err := SerializableShardValidatorListToValidatorList(shardValidators)
		if err != nil {
			return nil, err
		}
		shardIdInt, err := strconv.ParseUint(shardId, 10, 32)
		if err != nil {
			return nil, err
		}
		validators[uint32(shardIdInt)] = newValidators
	}

	return validators, nil
}

// SerializableShardValidatorListToValidatorList creates the validator list from serializable validator list
func SerializableShardValidatorListToValidatorList(shardValidators []*SerializableValidator) ([]Validator, error) {
	newValidators := make([]Validator, len(shardValidators))
	for i, shardValidator := range shardValidators {
		v, err := NewValidator(shardValidator.PubKey, shardValidator.Chances, shardValidator.Index)
		if err != nil {
			return nil, err
		}
		newValidators[i] = v
	}
	return newValidators, nil
}
