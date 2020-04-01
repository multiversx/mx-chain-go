package sharding

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

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

	log.Debug("selectValidators", "randomness", randomness, "validators", strValidators)
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
