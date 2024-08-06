package scToProtocol

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type modifiedMBShardIDChecker struct {
}

func (c *modifiedMBShardIDChecker) isModifiedStateMBValid(miniBlock *block.MiniBlock) bool {
	return miniBlock.SenderShardID == core.MetachainShardId && miniBlock.ReceiverShardID == core.MetachainShardId
}
