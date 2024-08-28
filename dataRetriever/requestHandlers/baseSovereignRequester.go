package requestHandlers

import "github.com/multiversx/mx-chain-go/dataRetriever"

type baseSovereignRequest struct {
	requestersFinder dataRetriever.RequestersFinder
}

func (br *baseSovereignRequest) getTrieNodeRequester(topic string) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("sovereignResolverRequestHandler.getTrieNodeRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

func (br *baseSovereignRequest) getTrieNodesRequester(topic string, _ uint32) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("sovereignResolverRequestHandler.getTrieNodesRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

func (br *baseSovereignRequest) getStartOfEpochMetaBlockRequester(topic string) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("sovereignResolverRequestHandler.getStartOfEpochMetaBlockRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}
