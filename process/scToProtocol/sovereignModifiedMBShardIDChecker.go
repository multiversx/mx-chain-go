package scToProtocol

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type sovModifiedMBShardIDChecker struct {
}

func (c *sovModifiedMBShardIDChecker) isModifiedStateMBValid(miniBlock *block.MiniBlock) bool {
	return miniBlock.SenderShardID == core.SovereignChainShardId && miniBlock.ReceiverShardID == core.SovereignChainShardId
}
