package containers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers"
)

type ResolverRequestHandler struct {
	resolversFinder dataRetriever.ResolversFinder
	txRequestTopic  string
	mbRequestTopic  string
	hdrRequestTopic string
	isMetaChain     bool
	maxTxsToRequest int
}

var log = logger.DefaultLogger()

func NewShardResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	txRequestTopic string,
	mbRequestTopic string,
	hdrRequestTopic string,
	maxTxsToRequest int,
) (*ResolverRequestHandler, error) {
	if finder == nil {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if len(txRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyTxRequestTopic
	}
	if len(mbRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyMiniBlockRequestTopic
	}
	if len(hdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyHeaderRequestTopic
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}

	rrh := &ResolverRequestHandler{
		txRequestTopic:  txRequestTopic,
		mbRequestTopic:  mbRequestTopic,
		hdrRequestTopic: hdrRequestTopic,
		isMetaChain:     false,
		maxTxsToRequest: maxTxsToRequest,
	}

	return rrh, nil
}

func NewMetaResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	hdrRequestTopic string,
) (*ResolverRequestHandler, error) {
	if finder == nil {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if len(hdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyHeaderRequestTopic
	}

	rrh := &ResolverRequestHandler{
		hdrRequestTopic: hdrRequestTopic,
		isMetaChain:     true,
	}

	return rrh, nil
}

func (rrh *ResolverRequestHandler) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	if len(rrh.txRequestTopic) == 0 {
		log.Error(fmt.Sprintf("transaction request topic is not set"))
		return
	}

	log.Debug(fmt.Sprintf("Requesting %d transactions from shard %d from network...\n", len(txHashes), destShardID))
	resolver, err := rrh.resolversFinder.CrossShardResolver(rrh.txRequestTopic, destShardID)
	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", rrh.txRequestTopic, destShardID))
		return
	}

	txResolver, ok := resolver.(*resolvers.TxResolver)
	if !ok {
		log.Error("wrong assertion type when creating transaction resolver")
		return
	}

	go func() {
		dataSplit := &partitioning.DataSplit{}
		sliceBatches, err := dataSplit.SplitDataInChunks(txHashes, rrh.maxTxsToRequest)
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

func (rrh *ResolverRequestHandler) RequestMiniBlock(shardId uint32, miniblockHash []byte) {
	if len(rrh.mbRequestTopic) == 0 {
		log.Error(fmt.Sprintf("miniblock request topic is not set"))
		return
	}

	rrh.requestByHash(shardId, miniblockHash, rrh.mbRequestTopic)
}

func (rrh *ResolverRequestHandler) RequestHeader(shardId uint32, hash []byte) {
	if len(rrh.hdrRequestTopic) == 0 {
		log.Error(fmt.Sprintf("header request topic is not set"))
		return
	}

	rrh.requestByHash(shardId, hash, rrh.hdrRequestTopic)
}

func (rrh *ResolverRequestHandler) requestByHash(destShardID uint32, hash []byte, baseTopic string) {
	log.Debug(fmt.Sprintf("Requesting %s from shard %d with hash %s from network\n", baseTopic, destShardID, core.ToB64(hash)))
	resolver, err := rrh.resolversFinder.CrossShardResolver(baseTopic, destShardID)
	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", baseTopic, destShardID))
		return
	}

	err = resolver.RequestDataFromHash(hash)
	if err != nil {
		log.Debug(err.Error())
	}
}

func (rrh *ResolverRequestHandler) RequestHeaderByNonce(destShardID uint32, nonce uint64) {
	if len(rrh.hdrRequestTopic) == 0 {
		log.Error(fmt.Sprintf("header by nonce request topic is not set"))
		return
	}

	if rrh.isMetaChain {
		rrh.requestWithCrossShardResolver(destShardID, nonce)
		return
	} else {
		rrh.requestWithMetachainResolver(destShardID, nonce)
		return
	}
}

func (rrh *ResolverRequestHandler) requestWithCrossShardResolver(destShardID uint32, nonce uint64) {
	log.Debug(fmt.Sprintf("Requesting %s from shard %d with nonce %d from network\n", rrh.hdrRequestTopic, destShardID, nonce))
	resolver, err := rrh.resolversFinder.CrossShardResolver(rrh.hdrRequestTopic, destShardID)
	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
		return
	}

	headerResolver, ok := resolver.(*resolvers.ShardHeaderResolver)
	if !ok {
		log.Error(fmt.Sprintf("resolver is not a header resolver to %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
		return
	}

	err = headerResolver.RequestDataFromNonce(nonce)
	if err != nil {
		log.Debug(err.Error())
	}
}

func (rrh *ResolverRequestHandler) requestWithMetachainResolver(destShardID uint32, nonce uint64) {
	log.Debug(fmt.Sprintf("Requesting %s from shard %d with nonce %d from network\n", rrh.hdrRequestTopic, destShardID, nonce))
	resolver, err := rrh.resolversFinder.MetaChainResolver(rrh.hdrRequestTopic)
	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
		return
	}

	headerResolver, ok := resolver.(*resolvers.HeaderResolver)
	if !ok {
		log.Error(fmt.Sprintf("resolver is not a header resolverto %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
		return
	}

	err = headerResolver.RequestDataFromNonce(nonce)
	if err != nil {
		log.Debug(err.Error())
	}
}
