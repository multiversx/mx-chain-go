package spos

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// GetConsensusTopicIDFromShardCoordinator will construct and return the topic ID based on shard coordinator
func GetConsensusTopicIDFromShardCoordinator(shardCoord sharding.Coordinator) string {
	return core.ConsensusTopic +
		shardCoord.CommunicationIdentifier(shardCoord.SelfId())
}
