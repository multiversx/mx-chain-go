package sharding

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

func cloneValidatorsMap(validatorsMap map[uint32][]Validator) (map[uint32][]Validator, error) {
	var err error
	resultMap := make(map[uint32][]Validator)
	for shard, vList := range validatorsMap {
		resultMap[shard], err = cloneValidatorsList(vList)
		if err != nil {
			return nil, err
		}
	}
	return resultMap, nil
}

func cloneValidatorsList(validatorsList []Validator) ([]Validator, error) {
	resultList := make([]Validator, len(validatorsList))
	for i, v := range validatorsList {
		clone, err := v.Clone()
		if err != nil {
			return nil, err
		}
		resultList[i] = clone.(Validator)
	}
	return resultList, nil
}

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
		strValidators += "\n" + core.ToHex(v.PubKey())
	}

	log.Trace("selectValidators", "randomness", randomness, "validators", strValidators)
}

func displayNodesConfiguration(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	leaving []Validator,
	actualRemaining []Validator,
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
	}

	for _, v := range leaving {
		pk := v.PubKey()
		log.Debug("computed leaving", "pk", pk)
	}

	for _, v := range actualRemaining {
		pk := v.PubKey()
		log.Debug("actually remaining", "pk", pk)
	}
}
