package spos

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// GetConsensusTopicID will construct and return the topic ID based on shard coordinator
func GetConsensusTopicID(shardCoordinator sharding.Coordinator) string {
	return core.ConsensusTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())
}
