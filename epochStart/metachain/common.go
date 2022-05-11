package metachain

import "github.com/ElrondNetwork/elrond-go/state"

// GetAllNodeKeys returns all <shard,pubKeys> from the provided man
func GetAllNodeKeys(validatorsInfo state.ShardValidatorsInfoMapHandler) map[uint32][][]byte {
	nodeKeys := make(map[uint32][][]byte)
	for shardID, validatorsInfoSlice := range validatorsInfo.GetShardValidatorsInfoMap() {
		nodeKeys[shardID] = make([][]byte, 0)
		for _, validatorInfo := range validatorsInfoSlice {
			nodeKeys[shardID] = append(nodeKeys[shardID], validatorInfo.GetPublicKey())
		}
	}

	return nodeKeys
}
