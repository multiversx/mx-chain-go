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
	resolversFinder      dataRetriever.ResolversFinder
	txRequestTopic       string
	scrRequestTopic      string
	rewardTxRequestTopic string
	mbRequestTopic       string
	hdrRequestTopic      string
	isMetaChain          bool
	maxTxsToRequest      int
}

var log = logger.DefaultLogger()

// NewShardResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewShardResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	txRequestTopic string,
	scrRequestTopic string,
	rewardTxRequestTopic string,
	mbRequestTopic string,
	hdrRequestTopic string,
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
	if len(hdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyHeaderRequestTopic
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}

	rrh := &resolverRequestHandler{
		resolversFinder:      finder,
		txRequestTopic:       txRequestTopic,
		mbRequestTopic:       mbRequestTopic,
		hdrRequestTopic:      hdrRequestTopic,
		scrRequestTopic:      scrRequestTopic,
		rewardTxRequestTopic: rewardTxRequestTopic,
		isMetaChain:          false,
		maxTxsToRequest:      maxTxsToRequest,
	}

	return rrh, nil
}

// NewMetaResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewMetaResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	hdrRequestTopic string,
) (*resolverRequestHandler, error) {
	if finder == nil || finder.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if len(hdrRequestTopic) == 0 {
		return nil, dataRetriever.ErrEmptyHeaderRequestTopic
	}

	rrh := &resolverRequestHandler{
		resolversFinder: finder,
		hdrRequestTopic: hdrRequestTopic,
		isMetaChain:     true,
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
func (rrh *resolverRequestHandler) RequestRewardTransactions(destShardId uint32, rewardTxHashes [][]byte){
	rrh.requestByHashes(destShardId, rewardTxHashes, rrh.rewardTxRequestTopic)
}

// RequestMiniBlock method asks for miniblocks from the connected peers
func (rrh *resolverRequestHandler) RequestMiniBlock(shardId uint32, miniblockHash []byte) {
	rrh.requestByHash(shardId, miniblockHash, rrh.mbRequestTopic)
}

// RequestHeader method asks for header from the connected peers
func (rrh *resolverRequestHandler) RequestHeader(shardId uint32, hash []byte) {
	rrh.requestByHash(shardId, hash, rrh.hdrRequestTopic)
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
	if rrh.isMetaChain {
		resolver, err = rrh.resolversFinder.CrossShardResolver(rrh.hdrRequestTopic, destShardID)
	} else {
		resolver, err = rrh.resolversFinder.MetaChainResolver(rrh.hdrRequestTopic)
	}

	if err != nil {
		log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
		return
	}

	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		log.Error(fmt.Sprintf("resolver is not a header resolver to %s topic to shard %d", rrh.hdrRequestTopic, destShardID))
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
