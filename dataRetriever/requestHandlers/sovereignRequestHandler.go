package requestHandlers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
)

const sovUniqueHeadersSuffix = "sovHdr"

type sovereignResolverRequestHandler struct {
	*resolverRequestHandler
}

// NewSovereignResolverRequestHandler creates a sovereignRequestHandler interface implementation with request functions
func NewSovereignResolverRequestHandler(resolverRequestHandler *resolverRequestHandler) (*sovereignResolverRequestHandler, error) {
	if resolverRequestHandler == nil {
		return nil, process.ErrNilRequestHandler
	}

	srrh := &sovereignResolverRequestHandler{
		resolverRequestHandler,
	}

	srrh.shardID = core.SovereignChainShardId
	return srrh, nil
}

// RequestExtendedShardHeaderByNonce method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeaderByNonce(nonce uint64) {
	suffix := fmt.Sprintf("%s_%d", sovUniqueHeadersSuffix, srrh.shardID)
	key := []byte(fmt.Sprintf("%d-%d", srrh.shardID, nonce))
	if !srrh.testIfRequestIsNeeded(key, suffix) {
		return
	}

	log.Debug("RequestExtendedShardHeaderByNonce.getExtendedShardHeaderRequester: requesting extended shard header by nonce from network",
		"shard", srrh.shardID,
		"nonce", nonce,
	)

	requester, err := srrh.getExtendedShardHeaderRequester()
	if err != nil {
		log.Error("RequestExtendedShardHeaderByNonce.getExtendedShardHeaderRequester",
			"error", err.Error(),
			"shard", srrh.shardID,
		)
		return
	}

	headerRequester, ok := requester.(NonceRequester)
	if !ok {
		log.Error("sovereignResolverRequestHandler.RequestExtendedShardHeaderByNonce: wrong assertion type when creating header requester")
		return
	}

	srrh.whiteList.Add([][]byte{key})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromNonce(nonce, epoch)
	if err != nil {
		log.Debug("RequestExtendedShardHeaderByNonce.RequestDataFromNonce",
			"error", err.Error(),
			"epoch", epoch,
			"nonce", nonce,
		)
		return
	}

	srrh.addRequestedItems([][]byte{key}, suffix)
}

func (srrh *sovereignResolverRequestHandler) getExtendedShardHeaderRequester() (dataRetriever.Requester, error) {
	headerRequester, err := srrh.requestersFinder.IntraShardRequester(factory.ExtendedHeaderProofTopic)
	if err != nil {
		log.Warn("extended header proof container not found, available requesters in container",
			"requesters", srrh.requestersFinder.RequesterKeys(),
		)
		return nil, fmt.Errorf("%w, topic: %s", err, factory.ExtendedHeaderProofTopic)
	}

	return headerRequester, nil
}

// RequestExtendedShardHeader method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeader(hash []byte) {
	suffix := fmt.Sprintf("%s_%d", sovUniqueHeadersSuffix, srrh.shardID)
	if !srrh.testIfRequestIsNeeded(hash, suffix) {
		return
	}

	log.Debug("sovereignResolverRequestHandler.RequestExtendedShardHeader: requesting extended shard header from network by hash",
		"shard", srrh.shardID,
		"hash", hash,
	)

	headerRequester, err := srrh.getExtendedShardHeaderRequester()
	if err != nil {
		log.Error("RequestExtendedShardHeader.getExtendedShardHeaderRequester",
			"error", err.Error(),
			"shard", srrh.shardID,
		)
		return
	}
	// equesting header by epo
	srrh.whiteList.Add([][]byte{hash})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromHash(hash, epoch)
	if err != nil {
		log.Debug("RequestExtendedShardHeader.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", hash,
		)
		return
	}

	srrh.addRequestedItems([][]byte{hash}, suffix)
}

// RequestTrieNode method asks for a trie node from the connected peers by the hash and the chunk index
func (srrh *sovereignResolverRequestHandler) RequestTrieNode(requestHash []byte, topic string, chunkIndex uint32) {
	srrh.requestTrieNode(requestHash, topic, chunkIndex, srrh.getTrieNodeRequester)
}

func (srrh *sovereignResolverRequestHandler) getTrieNodeRequester(topic string) (dataRetriever.Requester, error) {
	requester, err := srrh.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("sovereignResolverRequestHandler.getTrieNodeRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

// RequestTrieNodes method asks for trie nodes from the connected peers
func (srrh *sovereignResolverRequestHandler) RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string) {
	srrh.requestTrieNodes(destShardID, hashes, topic, srrh.getTrieNodesRequester)
}

func (srrh *sovereignResolverRequestHandler) getTrieNodesRequester(topic string, _ uint32) (dataRetriever.Requester, error) {
	requester, err := srrh.requestersFinder.IntraShardRequester(topic)
	if err != nil {
		log.Error("sovereignResolverRequestHandler.getTrieNodesRequester.IntraShardRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return nil, err
	}

	return requester, nil
}

// RequestStartOfEpochMetaBlock method asks for the start of epoch metablock from the connected peers
func (srrh *sovereignResolverRequestHandler) RequestStartOfEpochMetaBlock(epoch uint32) {
	epochStartIdentifier := []byte(core.EpochStartIdentifier(epoch))
	if !srrh.testIfRequestIsNeeded(epochStartIdentifier, uniqueMetaHeadersSuffix) {
		return
	}

	baseTopic := factory.ShardBlocksTopic
	log.Debug("requesting header by epoch",
		"topic", baseTopic,
		"epoch", epoch,
		"hash", epochStartIdentifier,
	)

	requester, err := srrh.requestersFinder.IntraShardRequester(baseTopic)
	if err != nil {
		log.Error("RequestStartOfEpochMetaBlock.MetaChainRequester",
			"error", err.Error(),
			"topic", baseTopic,
		)
		return
	}

	headerRequester, ok := requester.(EpochRequester)
	if !ok {
		log.Warn("wrong assertion type when creating header requester")
		return
	}

	srrh.whiteList.Add([][]byte{epochStartIdentifier})

	err = headerRequester.RequestDataFromEpoch(epochStartIdentifier)
	if err != nil {
		log.Debug("RequestStartOfEpochMetaBlock.RequestDataFromEpoch",
			"error", err.Error(),
			"epochStartIdentifier", epochStartIdentifier,
		)
		return
	}

	srrh.addRequestedItems([][]byte{epochStartIdentifier}, uniqueMetaHeadersSuffix)
}

// RequestMetaHeader method asks for meta header from the connected peers
func (srrh *sovereignResolverRequestHandler) RequestMetaHeader(hash []byte) {
	if !srrh.testIfRequestIsNeeded(hash, uniqueMetaHeadersSuffix) {
		return
	}

	log.Debug("requesting meta header from network",
		"hash", hash,
	)

	requester, err := srrh.getMetaHeaderRequester()
	if err != nil {
		log.Error("RequestMetaHeader.getMetaHeaderRequester",
			"error", err.Error(),
			"hash", hash,
		)
		return
	}

	headerRequester, ok := requester.(dataRetriever.Requester)
	if !ok {
		log.Warn("wrong assertion type when creating header requester")
		return
	}

	srrh.whiteList.Add([][]byte{hash})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromHash(hash, epoch)
	if err != nil {
		log.Debug("RequestMetaHeader.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", hash,
		)
		return
	}

	srrh.addRequestedItems([][]byte{hash}, uniqueMetaHeadersSuffix)
}

func (srrh *sovereignResolverRequestHandler) getMetaHeaderRequester() (HeaderRequester, error) {
	requester, err := srrh.requestersFinder.IntraShardRequester(factory.ShardBlocksTopic)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d",
			err, factory.MetachainBlocksTopic, srrh.shardID)
		return nil, err
	}

	headerRequester, ok := requester.(HeaderRequester)
	if !ok {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, expected HeaderRequester",
			dataRetriever.ErrWrongTypeInContainer, factory.ShardBlocksTopic, srrh.shardID)
		return nil, err
	}

	return headerRequester, nil
}
