package process

import (
	"fmt"
)

func ShardCacherIdentifier(senderShardId uint32, destinationShardId uint32) string {
	if senderShardId == destinationShardId {
		return fmt.Sprintf("%d", senderShardId)
	}

	return fmt.Sprintf("%d_%d", senderShardId, destinationShardId)
}
