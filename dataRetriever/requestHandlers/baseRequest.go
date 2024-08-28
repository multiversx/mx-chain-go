package requestHandlers

import "github.com/multiversx/mx-chain-go/dataRetriever"

type baseRequest struct {
	requestersFinder dataRetriever.RequestersFinder
}

func (br *baseRequest) getTrieNodeRequester(topic string) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.MetaChainRequester(topic)
	if err != nil {
		log.Error("requestersFinder.MetaChainRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

func (br *baseRequest) getTrieNodesRequester(topic string, destShardID uint32) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.MetaCrossShardRequester(topic, destShardID)
	if err != nil {
		log.Error("requestersFinder.MetaCrossShardRequester",
			"error", err.Error(),
			"topic", topic,
			"shard", destShardID,
		)
		return nil, err
	}

	return requester, err
}

func (br *baseRequest) getStartOfEpochMetaBlockRequester(topic string) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.MetaChainRequester(topic)
	if err != nil {
		log.Error("RequestStartOfEpochMetaBlock.MetaChainRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}
