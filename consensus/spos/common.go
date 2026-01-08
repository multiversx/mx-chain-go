package spos

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/sharding"
)

// GetConsensusTopicID will construct and return the topic ID based on shard coordinator
func GetConsensusTopicID(shardCoordinator sharding.Coordinator) string {
	return common.ConsensusTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())
}

// ShouldConsiderSelfKeyInConsensus returns true if current machine is the main one, or it is a backup machine but the main
// machine failed
func ShouldConsiderSelfKeyInConsensus(redundancyHandler consensus.NodeRedundancyHandler) bool {
	if check.IfNil(redundancyHandler) {
		log.Warn("redundancy handler is nil")
		return false
	}

	isMainMachine := !redundancyHandler.IsRedundancyNode()
	if isMainMachine {
		return true
	}
	isMainMachineInactive := !redundancyHandler.IsMainMachineActive()

	return isMainMachineInactive
}
