package bootstrap

import "github.com/ElrondNetwork/elrond-go/data/block"

func (e *epochStartMetaBlockProcessor) GetMapMetaBlock() map[string]*block.MetaBlock {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	return e.mapReceivedMetaBlocks
}

const DurationBetweenChecksForEpochStartMetaBlock = durationBetweenChecks

const DurationBetweenReRequest = durationBetweenReRequests
