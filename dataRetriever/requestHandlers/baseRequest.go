package requestHandlers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
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

func (br *baseRequest) getShardHeaderRequester(shardID uint32) (dataRetriever.Requester, error) {
	isMetachainNode := br.shardID == core.MetachainShardId
	shardIdMissmatch := br.shardID != shardID
	requestOnMetachain := shardID == core.MetachainShardId
	isRequestInvalid := (!isMetachainNode && shardIdMissmatch) || requestOnMetachain
	if isRequestInvalid {
		return nil, dataRetriever.ErrBadRequest
	}

	// requests should be done on the topic shardBlocks_0_META so that is why we need to figure out
	// the cross shard id
	crossShardID := core.MetachainShardId
	if isMetachainNode {
		crossShardID = shardID
	}

	headerRequester, err := br.requestersFinder.CrossShardRequester(factory.ShardBlocksTopic, crossShardID)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, cross shard ID: %d",
			err, factory.ShardBlocksTopic, br.shardID, crossShardID)

		log.Warn("available requesters in container",
			"requesters", br.requestersFinder.RequesterKeys(),
		)
		return nil, err
	}

	return headerRequester, nil
}

func (br *baseRequest) getValidatorsInfoRequester() (dataRetriever.Requester, error) {
	return br.requestersFinder.MetaChainRequester(common.ValidatorInfoTopic)
}

func (br *baseRequest) getMiniBlocksRequester(destShardID uint32) (dataRetriever.Requester, error) {
	return br.requestersFinder.CrossShardRequester(factory.MiniBlocksTopic, destShardID)
}
