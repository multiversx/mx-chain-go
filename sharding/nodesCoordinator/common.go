package nodesCoordinator

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("nodesCoordinator")

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
	if log.GetLevel() != logger.LogTrace {
		return
	}

	strValidators := ""

	for _, v := range validators {
		strValidators += "\n" + hex.EncodeToString(v.PubKey())
	}

	log.Trace("selectValidators", "randomness", randomness, "validators", strValidators)
}

func DisplayNodesConfiguration(
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

// CopyValidatorMap creates a copy for the Validators map, creating copies for each of the lists for each shard
func CopyValidatorMap(validatorsMap map[uint32][]Validator) map[uint32][]Validator {
	result := make(map[uint32][]Validator)

	for shardId, validators := range validatorsMap {
		elems := make([]Validator, 0)
		result[shardId] = append(elems, validators...)
	}

	return result
}
