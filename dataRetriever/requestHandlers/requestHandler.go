package requestHandlers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type ResolverRequestHandler struct {
	resolversFinder      dataRetriever.ResolversFinder
	txRequestTopic       string
	scrRequestTopic      string
	mbRequestTopic       string
	shardHdrRequestTopic string
	metaHdrRequestTopic  string
	isMetaChain          bool
	maxTxsToRequest      int
}

var log = logger.DefaultLogger()

// NewShardResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewShardResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	txRequestTopic string,
	scrRequestTopic string,
	mbRequestTopic string,
	shardHdrRequestTopic string,
	metaHdrRequestTopic string,
	maxTxsToRequest int,
) (*ResolverRequestHandler, error) {
	if finder == nil || finder.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverFinder
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
	if len(shardHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyShardHeaderRequestTopic
	}
	if len(metaHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMetaHeaderRequestTopic
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}

	rrh := &ResolverRequestHandler{
		resolversFinder:      finder,
		txRequestTopic:       txRequestTopic,
		mbRequestTopic:       mbRequestTopic,
		shardHdrRequestTopic: shardHdrRequestTopic,
		metaHdrRequestTopic:  metaHdrRequestTopic,
		scrRequestTopic:      scrRequestTopic,
		isMetaChain:          false,
		maxTxsToRequest:      maxTxsToRequest,
	}

	return rrh, nil
}

// NewMetaResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewMetaResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	shardHdrRequestTopic string,
	metaHdrRequestTopic string,
) (*ResolverRequestHandler, error) {
	if finder == nil || finder.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if len(shardHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyShardHeaderRequestTopic
	}
	if len(metaHdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMetaHeaderRequestTopic
	}

	rrh := &ResolverRequestHandler{
		resolversFinder:      finder,
		shardHdrRequestTopic: shardHdrRequestTopic,
		metaHdrRequestTopic:  metaHdrRequestTopic,
		isMetaChain:          true,
	}

	return rrh, nil
}

// RequestTransaction method asks for transactions from the connected peers
func (rrh *ResolverRequestHandler) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	rrh.requestByHashes(destShardID, txHashes, rrh.txRequestTopic)
}

func (rrh *ResolverRequestHandler) requestByHashes(destShardID uint32, hashes [][]byte, topic string) {
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
func (rrh *ResolverRequestHandler) RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte) {
	rrh.requestByHashes(destShardID, scrHashes, rrh.scrRequestTopic)
}

// RequestMiniBlock method asks for miniblocks from the connected peers
func (rrh *ResolverRequestHandler) RequestMiniBlock(shardId uint32, miniblockHash []byte) {
	rrh.requestByHash(shardId, miniblockHash, rrh.mbRequestTopic)
}

// RequestHeader method asks for header from the connected peers
func (rrh *ResolverRequestHandler) RequestHeader(shardId uint32, hash []byte) {
	var topic string
	if shardId == sharding.MetachainShardId {
		topic = rrh.metaHdrRequestTopic
	} else {
		topic = rrh.shardHdrRequestTopic
	}

	rrh.requestByHash(shardId, hash, topic)
}

func (rrh *ResolverRequestHandler) requestByHash(destShardID uint32, hash []byte, baseTopic string) {
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
func (rrh *ResolverRequestHandler) RequestHeaderByNonce(destShardID uint32, nonce uint64) {
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
func (rrh *ResolverRequestHandler) IsInterfaceNil() bool {
	if rrh == nil {
		return true
	}
	return false
}
