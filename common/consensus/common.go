package consensus

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("common/consensus")

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
	return !redundancyHandler.IsMainMachineActive()
}
