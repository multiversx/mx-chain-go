package spos

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
)

// GetConsensusTopicID will construct and return the topic ID based on shard coordinator
func GetConsensusTopicID(shardCoordinator consensus.ShardCoordinator) string {
	return consensus.ConsensusTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())
}
