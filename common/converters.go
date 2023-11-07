package common

import (
	"fmt"
	"strconv"
	"strings"
)

// ProcessDestinationShardAsObserver returns the shardID given the destination as observer string
func ProcessDestinationShardAsObserver(destinationShardIdAsObserver string) (uint32, error) {
	destShard := strings.ToLower(destinationShardIdAsObserver)
	if len(destShard) == 0 {
		return 0, fmt.Errorf("option DestinationShardAsObserver is not set in prefs.toml")
	}

	if destShard == NotSetDestinationShardID {
		return DisabledShardIDAsObserver, nil
	}

	if destShard == MetachainShardName {
		return MetachainShardId, nil
	}

	val, err := strconv.ParseUint(destShard, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("error parsing DestinationShardAsObserver option: " + err.Error())
	}

	return uint32(val), err
}

// AssignShardForPubKeyWhenNotSpecified will return the same shard ID when it is called with the same parameters
// This function fetched the last byte of the public key and based on a modulo operation it will return a shard ID
func AssignShardForPubKeyWhenNotSpecified(pubKey []byte, numShards uint32) uint32 {
	if len(pubKey) == 0 {
		return 0 // should never happen
	}

	lastByte := pubKey[len(pubKey)-1]
	numShardsIncludingMeta := numShards + 1

	randomShardID := uint32(lastByte) % numShardsIncludingMeta
	if randomShardID == numShards {
		randomShardID = MetachainShardId
	}

	return randomShardID
}

// SuffixedMetric appends the suffix to the provided metric and returns it
func SuffixedMetric(metric string, suffix string) string {
	return fmt.Sprintf("%s%s", metric, suffix)
}
