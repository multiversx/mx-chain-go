package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignTopicProvider struct {
}

// internal constructor
func newSovereignTopicProvider() *sovereignTopicProvider {
	return &sovereignTopicProvider{}
}

func (e *sovereignTopicProvider) getTopic() string {
	return factory.ShardBlocksTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
}

type epochStartSovereignBlockProcessor struct {
	*sovereignTopicProvider
	*epochStartMetaBlockProcessor
}

// internal constructor
func newEpochStartSovereignBlockProcessor(epochStartMetaBlockProcessor *epochStartMetaBlockProcessor) *epochStartSovereignBlockProcessor {
	sovBlockProc := &epochStartSovereignBlockProcessor{
		epochStartMetaBlockProcessor: epochStartMetaBlockProcessor,
		sovereignTopicProvider:       &sovereignTopicProvider{},
	}

	sovBlockProc.epochStartPeerHandler = sovBlockProc
	return sovBlockProc
}

func (e *epochStartSovereignBlockProcessor) setNumPeers(
	requestHandler RequestHandler,
	intra int, _ int,
) error {
	return requestHandler.SetNumPeersToQuery(e.sovereignTopicProvider.getTopic(), intra, 0)
}

func (e *epochStartSovereignBlockProcessor) getTopic() string {
	return e.sovereignTopicProvider.getTopic()
}
