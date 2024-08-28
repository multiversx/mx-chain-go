package requestHandlers

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type baseRequest struct {
	shardID          uint32
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

func (br *baseRequest) getMetaHeaderRequester() (HeaderRequester, error) {
	requester, err := br.requestersFinder.MetaChainRequester(factory.MetachainBlocksTopic)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d",
			err, factory.MetachainBlocksTopic, br.shardID)
		return nil, err
	}

	headerRequester, ok := requester.(HeaderRequester)
	if !ok {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, expected HeaderRequester",
			dataRetriever.ErrWrongTypeInContainer, factory.ShardBlocksTopic, br.shardID)
		return nil, err
	}

	return headerRequester, nil
}
