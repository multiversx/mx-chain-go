package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type epochStartSovereignBlockProcessor struct {
	*epochStartMetaBlockProcessor
}

// internal constructor
func newEpochStartSovereignBlockProcessor(epochStartMetaBlockProcessor *epochStartMetaBlockProcessor) *epochStartSovereignBlockProcessor {
	sovBlockProc := &epochStartSovereignBlockProcessor{
		epochStartMetaBlockProcessor,
	}

	sovBlockProc.epochStartPeerHandler = sovBlockProc
	return sovBlockProc
}

func (e *epochStartSovereignBlockProcessor) setNumPeers(
	requestHandler RequestHandler,
	intra int, _ int,
) error {
	return requestHandler.SetNumPeersToQuery(e.getRequestTopic(), intra, 0)
}

func (e *epochStartSovereignBlockProcessor) getRequestTopic() string {
	return factory.ShardBlocksTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
}
