package requestHandlers

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type baseSovereignRequest struct {
	requestersFinder dataRetriever.RequestersFinder
}

func (br *baseSovereignRequest) getTrieNodeRequester(topic string) (dataRetriever.Requester, error) {
	requester, err := br.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("baseSovereignRequest.getTrieNodeRequester.IntraShardRequester",
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
		log.Error("baseSovereignRequest.getTrieNodesRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

func (br *baseSovereignRequest) getStartOfEpochMetaBlockRequester(_ string) (dataRetriever.Requester, error) {
	topic := factory.ShardBlocksTopic
	log.Debug("baseSovereignRequest.getStartOfEpochMetaBlockRequester", "topic", topic)

	requester, err := br.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("baseSovereignRequest.getStartOfEpochMetaBlockRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

func (br *baseSovereignRequest) getMetaHeaderRequester() (HeaderRequester, error) {
	requester, err := br.requestersFinder.IntraShardRequester(factory.ShardBlocksTopic)
	if err != nil {
		err = fmt.Errorf("baseSovereignRequest.getMetaHeaderRequester: %w, topic: %s",
			err, factory.ShardBlocksTopic)
		return nil, err
	}

	headerRequester, ok := requester.(HeaderRequester)
	if !ok {
		err = fmt.Errorf("baseSovereignRequest.getMetaHeaderRequester: %w, topic: %s, expected HeaderRequester",
			dataRetriever.ErrWrongTypeInContainer, factory.ShardBlocksTopic)
		return nil, err
	}

	return headerRequester, nil
}

func (br *baseSovereignRequest) getShardHeaderRequester(_ uint32) (dataRetriever.Requester, error) {
	headerRequester, err := br.requestersFinder.IntraShardRequester(factory.ShardBlocksTopic)
	if err != nil {
		err = fmt.Errorf("baseSovereignRequest.getShardHeaderRequester: %w, topic: %s",
			err, factory.ShardBlocksTopic)

		log.Warn("available requesters in container",
			"requesters", br.requestersFinder.RequesterKeys(),
		)
		return nil, err
	}

	return headerRequester, nil
}

func (br *baseSovereignRequest) getValidatorsInfoRequester() (dataRetriever.Requester, error) {
	return br.requestersFinder.IntraShardRequester(common.ValidatorInfoTopic)
}

func (br *baseSovereignRequest) getMiniBlocksRequester(_ uint32) (dataRetriever.Requester, error) {
	return br.requestersFinder.IntraShardRequester(factory.MiniBlocksTopic)
}
