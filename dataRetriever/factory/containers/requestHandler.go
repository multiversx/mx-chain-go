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
}

const maxTxsToRequest = 100

var log = logger.DefaultLogger()

func NewShardResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	txRequestTopic string,
	mbRequestTopic string,
	hdrRequestTopic string,
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

	rrh := &ResolverRequestHandler{
		txRequestTopic:  txRequestTopic,
		mbRequestTopic:  mbRequestTopic,
		hdrRequestTopic: hdrRequestTopic,
		isMetaChain:     false,
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

func (rrh *ResolverRequestHandler) RequestTransactionHandler(destShardID uint32, txHashes [][]byte) {
	if len(rrh.txRequestTopic) == 0 {
		log.Error(fmt.Sprintf("not base topic for transaction request handler"))
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
		sliceBatches, err := dataSplit.SplitDataInChunks(txHashes, maxTxsToRequest)
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

func (rrh *ResolverRequestHandler) RequestMiniBlockHandler(shardId uint32, miniblockHash []byte) {
	if len(rrh.mbRequestTopic) == 0 {
		log.Error(fmt.Sprintf("not base topic for miniblock request handler"))
		return
	}

	rrh.requestHandlerByHash(shardId, miniblockHash, rrh.mbRequestTopic)
}

func (rrh *ResolverRequestHandler) RequestHeaderHandler(shardId uint32, hash []byte) {
	if len(rrh.hdrRequestTopic) == 0 {
		log.Error(fmt.Sprintf("not base topic for header request handler"))
		return
	}

	rrh.requestHandlerByHash(shardId, hash, rrh.hdrRequestTopic)
}

func (rrh *ResolverRequestHandler) requestHandlerByHash(destShardID uint32, hash []byte, baseTopic string) {
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

func (rrh *ResolverRequestHandler) RequestHeaderHandlerByNonce(destShardID uint32, nonce uint64) {
	if len(rrh.hdrRequestTopic) == 0 {
		log.Error(fmt.Sprintf("not base topic for header by nonce request handler"))
		return
	}

	if rrh.isMetaChain {
		log.Debug(fmt.Sprintf("Requesting %s from shard %d with nonce %d from network\n", rrh.hdrRequestTopic, destShardID, nonce))
		resolver, err := rrh.resolversFinder.CrossShardResolver(rrh.hdrRequestTopic, destShardID)
		if err != nil {
			log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
			return
		}

		headerResolver, ok := resolver.(*resolvers.ShardHeaderResolver)
		if !ok {
			log.Error(fmt.Sprintf("resolver is not a header resolverto %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
			return
		}

		err = headerResolver.RequestDataFromNonce(nonce)
		if err != nil {
			log.Debug(err.Error())
		}

		return
	} else {
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
}
