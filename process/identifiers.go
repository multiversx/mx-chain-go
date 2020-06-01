package process

import (
	"fmt"
	"strconv"
	"strings"
)

// ShardCacherIdentifier generates a string identifier between 2 shards
func ShardCacherIdentifier(senderShardID uint32, destinationShardID uint32) string {
	if senderShardID == destinationShardID {
		return fmt.Sprintf("%d", senderShardID)
	}

	return fmt.Sprintf("%d_%d", senderShardID, destinationShardID)
}

// ParseShardCacherIdentifier parses an identifier into its components: sender shard, destination shard
func ParseShardCacherIdentifier(cacheID string) (uint32, uint32, error) {
	parts := strings.Split(cacheID, "_")
	isIntraShard := len(parts) == 1
	isCrossShard := len(parts) == 2
	isBad := !isIntraShard && !isCrossShard

	if isBad {
		return parseErrorOfShardCacherIdentifier()
	}

	if isIntraShard {
		shardID, err := parseCacherIdentifierPart(parts[0])
		if err != nil {
			return parseErrorOfShardCacherIdentifier()
		}

		return shardID, shardID, nil
	}

	// Is cross-shard
	sourceShardID, err := parseCacherIdentifierPart(parts[0])
	if err != nil {
		return parseErrorOfShardCacherIdentifier()
	}

	destinationShardID, err := parseCacherIdentifierPart(parts[1])
	if err != nil {
		return parseErrorOfShardCacherIdentifier()
	}

	return sourceShardID, destinationShardID, nil
}

func parseErrorOfShardCacherIdentifier() (uint32, uint32, error) {
	return 0, 0, ErrInvalidShardCacherIdentifier
}

func parseCacherIdentifierPart(cacheID string) (uint32, error) {
	part, err := strconv.ParseUint(cacheID, 10, 64)
	return uint32(part), err
}

// IsShardCacherIdentifierForSourceMe checks whether the specified cache is for the pair (sender = me, destination = *)
func IsShardCacherIdentifierForSourceMe(cacheID string, shardID uint32) bool {
	parsed, err := parseCacherIdentifierPart(cacheID)
	if err != nil {
		return false
	}

	return parsed == shardID
}
