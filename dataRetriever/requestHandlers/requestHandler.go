package requestHandlers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type resolverRequestHandler struct {
	resolversFinder       dataRetriever.ResolversFinder
	txRequestTopic        string
	scrRequestTopic       string
	rewardTxRequestTopic  string
	mbRequestTopic        string
	shardHdrRequestTopic  string
	metaHdrRequestTopic   string
	trieNodesRequestTopic string
	isMetaChain           bool
	maxTxsToRequest       int
}

var log = logger.DefaultLogger()

// NewShardResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewShardResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	txRequestTopic string,
	scrRequestTopic string,
	rewardTxRequestTopic string,
	mbRequestTopic string,
	shardHdrRequestTopic string,
	metaHdrRequestTopic string,
	trieNodesRequestTopic string,
	maxTxsToRequest int,
) (*resolverRequestHandler, error) {
	if finder == nil || finder.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if len(txRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyTxRequestTopic
	}
	if len(scrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyScrRequestTopic
	}
	if len(rewardTxRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyRewardTxRequestTopic
	}
	if len(mbRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMiniBlockRequestTopic
	}
	if len(shardHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyShardHeaderRequestTopic
	}
	if len(metaHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMetaHeaderRequestTopic
	}
	if len(trieNodesRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyTrieNodesRequestTopic
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}

	rrh := &resolverRequestHandler{
		resolversFinder:       finder,
		txRequestTopic:        txRequestTopic,
		mbRequestTopic:        mbRequestTopic,
		shardHdrRequestTopic:  shardHdrRequestTopic,
		metaHdrRequestTopic:   metaHdrRequestTopic,
		scrRequestTopic:       scrRequestTopic,
		rewardTxRequestTopic:  rewardTxRequestTopic,
		trieNodesRequestTopic: trieNodesRequestTopic,
		isMetaChain:           false,
		maxTxsToRequest:       maxTxsToRequest,
	}

	return rrh, nil
}

// NewMetaResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewMetaResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	shardHdrRequestTopic string,
	metaHdrRequestTopic string,
	txRequestTopic string,
	scrRequestTopic string,
	mbRequestTopic string,
) (*resolverRequestHandler, error) {
	if finder == nil || finder.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if len(shardHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyShardHeaderRequestTopic
	}
	if len(metaHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMetaHeaderRequestTopic
	}
	if len(txRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyTxRequestTopic
	}
	if len(scrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyScrRequestTopic
	}
	if len(mbRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMiniBlockRequestTopic
	}

	rrh := &resolverRequestHandler{
		resolversFinder:      finder,
		shardHdrRequestTopic: shardHdrRequestTopic,
		metaHdrRequestTopic:  metaHdrRequestTopic,
		txRequestTopic:       txRequestTopic,
		mbRequestTopic:       mbRequestTopic,
		scrRequestTopic:      scrRequestTopic,
		isMetaChain:          true,
	}

	return rrh, nil
}

// RequestTransaction method asks for transactions from the connected peers
func (rrh *resolverRequestHandler) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	rrh.requestByHashes(destShardID, txHashes, rrh.txRequestTopic)
}

func (rrh *resolverRequestHandler) requestByHashes(destShardID uint32, hashes [][]byte, topic string) {
	log.Debug(fmt.Sprintf("Requesting %d transactions from shard %d from network on topic %s...\n", len(hashes), destShardID, topic))
	resolver, err := rrh.resolversFinder.CrossShardResolver(topic, destShardID)
	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", topic, destShardID))
		return
	}

	txResolver, ok := resolver.(HashSliceResolver)
	if !ok {
		log.Error("wrong assertion type when creating transaction resolver")
		return
	}

	go func() {
		dataSplit := &partitioning.DataSplit{}
		sliceBatches, err := dataSplit.SplitDataInChunks(hashes, rrh.maxTxsToRequest)
		if err != nil {
			log.Error("error requesting transactions: " + err.Error())
			return
		}

		for _, batch := range sliceBatches {
			err = txResolver.RequestDataFromHashArray(batch)
			if err != nil {
				log.Debug("error requesting tx batch: " + err.Error())
			}
		}
	}()
}

// RequestUnsignedTransactions method asks for unsigned transactions from the connected peers
func (rrh *resolverRequestHandler) RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte) {
	rrh.requestByHashes(destShardID, scrHashes, rrh.scrRequestTopic)
}

// RequestRewardTransactions requests for reward transactions from the connected peers
func (rrh *resolverRequestHandler) RequestRewardTransactions(destShardId uint32, rewardTxHashes [][]byte) {
	rrh.requestByHashes(destShardId, rewardTxHashes, rrh.rewardTxRequestTopic)
}

// RequestMiniBlock method asks for miniblocks from the connected peers
func (rrh *resolverRequestHandler) RequestMiniBlock(shardId uint32, miniblockHash []byte) {
	rrh.requestByHash(shardId, miniblockHash, rrh.mbRequestTopic)
}

// RequestHeader method asks for header from the connected peers
func (rrh *resolverRequestHandler) RequestHeader(shardId uint32, hash []byte) {
	//TODO: Refactor this class and create specific methods for requesting shard or meta data
	var topic string
	if shardId == sharding.MetachainShardId {
		topic = rrh.metaHdrRequestTopic
	} else {
		topic = rrh.shardHdrRequestTopic
	}

	rrh.requestByHash(shardId, hash, topic)
}

// RequestTrieNode method asks for trie nodes from the connected peers
func (rrh *resolverRequestHandler) RequestTrieNode(shardId uint32, hash []byte) {
	rrh.requestByHash(shardId, hash, rrh.trieNodesRequestTopic)
}

func (rrh *resolverRequestHandler) requestByHash(destShardID uint32, hash []byte, baseTopic string) {
	log.Debug(fmt.Sprintf("Requesting %s from shard %d with hash %s from network\n", baseTopic, destShardID, core.ToB64(hash)))

	var resolver dataRetriever.Resolver
	var err error

	if destShardID == sharding.MetachainShardId {
		resolver, err = rrh.resolversFinder.MetaChainResolver(baseTopic)
	} else {
		resolver, err = rrh.resolversFinder.CrossShardResolver(baseTopic, destShardID)
	}

	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", baseTopic, destShardID))
		return
	}

	err = resolver.RequestDataFromHash(hash)
	if err != nil {
		log.Debug(err.Error())
	}
}

// RequestHeaderByNonce method asks for transactions from the connected peers
func (rrh *resolverRequestHandler) RequestHeaderByNonce(destShardID uint32, nonce uint64) {
	var err error
	var resolver dataRetriever.Resolver
	var topic string
	if rrh.isMetaChain {
		topic = rrh.shardHdrRequestTopic
		resolver, err = rrh.resolversFinder.CrossShardResolver(topic, destShardID)
	} else {
		topic = rrh.metaHdrRequestTopic
		resolver, err = rrh.resolversFinder.MetaChainResolver(topic)
	}

	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", topic, destShardID))
		return
	}

	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		log.Error(fmt.Sprintf("resolver is not a header resolver to %s topic to shard %d", topic, destShardID))
		return
	}

	err = headerResolver.RequestDataFromNonce(nonce)
	if err != nil {
		log.Debug(err.Error())
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *resolverRequestHandler) IsInterfaceNil() bool {
	if rrh == nil {
		return true
	}
	return false
}
