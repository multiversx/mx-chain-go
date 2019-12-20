package requestHandlers

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type resolverRequestHandler struct {
	resolversFinder       dataRetriever.ResolversFinder
	requestedItemsHandler dataRetriever.RequestedItemsHandler
	shardID               uint32
	maxTxsToRequest       int
	sweepTime             time.Time
}

var log = logger.GetOrCreate("dataretriever/requesthandlers")

// NewShardResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewShardResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	maxTxsToRequest int,
	shardID uint32,
) (*resolverRequestHandler, error) {

	if check.IfNil(finder) {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if check.IfNil(requestedItemsHandler) {
		return nil, dataRetriever.ErrNilRequestedItemsHandler
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}

	rrh := &resolverRequestHandler{
		resolversFinder:       finder,
		requestedItemsHandler: requestedItemsHandler,
		shardID:               shardID,
		maxTxsToRequest:       maxTxsToRequest,
	}

	rrh.sweepTime = time.Now()

	return rrh, nil
}

// NewMetaResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewMetaResolverRequestHandler(
	finder dataRetriever.ResolversFinder,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	maxTxsToRequest int,
) (*resolverRequestHandler, error) {

	if check.IfNil(finder) {
		return nil, dataRetriever.ErrNilResolverFinder
	}
	if check.IfNil(requestedItemsHandler) {
		return nil, dataRetriever.ErrNilRequestedItemsHandler
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}

	rrh := &resolverRequestHandler{
		resolversFinder:       finder,
		requestedItemsHandler: requestedItemsHandler,
		shardID:               sharding.MetachainShardId,
		maxTxsToRequest:       maxTxsToRequest,
	}

	return rrh, nil
}

// RequestTransaction method asks for transactions from the connected peers
func (rrh *resolverRequestHandler) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	rrh.requestByHashes(destShardID, txHashes, factory.TransactionTopic)
}

func (rrh *resolverRequestHandler) requestByHashes(destShardID uint32, hashes [][]byte, topic string) {
	unrequestedHashes := rrh.getUnrequestedHashes(hashes)
	log.Trace("requesting transactions from network",
		"num txs", len(unrequestedHashes),
		"topic", topic,
		"shard", destShardID,
	)
	resolver, err := rrh.resolversFinder.CrossShardResolver(topic, destShardID)
	if err != nil {
		log.Error("missing resolver",
			"topic", topic,
			"shard", destShardID,
		)
		return
	}

	txResolver, ok := resolver.(HashSliceResolver)
	if !ok {
		log.Warn("wrong assertion type when creating transaction resolver")
		return
	}

	go func() {
		dataSplit := &partitioning.DataSplit{}
		sliceBatches, err := dataSplit.SplitDataInChunks(unrequestedHashes, rrh.maxTxsToRequest)
		if err != nil {
			log.Debug("requesting transactions", "error", err.Error())
			return
		}

		for _, batch := range sliceBatches {
			err = txResolver.RequestDataFromHashArray(batch)
			if err != nil {
				log.Debug("requesting tx batch", "error", err.Error())
			}
		}
	}()
}

// RequestUnsignedTransactions method asks for unsigned transactions from the connected peers
func (rrh *resolverRequestHandler) RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte) {
	rrh.requestByHashes(destShardID, scrHashes, factory.UnsignedTransactionTopic)
}

// RequestRewardTransactions requests for reward transactions from the connected peers
func (rrh *resolverRequestHandler) RequestRewardTransactions(destShardId uint32, rewardTxHashes [][]byte) {
	rrh.requestByHashes(destShardId, rewardTxHashes, factory.RewardsTransactionTopic)
}

// RequestMiniBlock method asks for miniblocks from the connected peers
func (rrh *resolverRequestHandler) RequestMiniBlock(destShardID uint32, miniblockHash []byte) {
	rrh.sweepIfNeeded()

	if rrh.requestedItemsHandler.Has(string(miniblockHash)) {
		log.Trace("item already requested",
			"key", miniblockHash)
		return
	}

	log.Trace("requesting miniblock from network",
		"hash", miniblockHash,
		"shard", destShardID,
		"topic", factory.MiniBlocksTopic,
	)

	resolver, err := rrh.resolversFinder.CrossShardResolver(factory.MiniBlocksTopic, destShardID)
	if err != nil {
		log.Error("missing resolver",
			"topic", factory.MiniBlocksTopic,
			"shard", destShardID,
		)
		return
	}

	err = resolver.RequestDataFromHash(miniblockHash)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	err = rrh.requestedItemsHandler.Add(string(miniblockHash))
	if err != nil {
		log.Trace("add requested item with error",
			"error", err.Error(),
			"key", miniblockHash)
	}
}

// RequestShardHeader method asks for shard header from the connected peers
func (rrh *resolverRequestHandler) RequestShardHeader(shardId uint32, hash []byte) {
	if !rrh.testIfRequestIsNeeded(hash) {
		return
	}

	headerResolver, err := rrh.getShardHeaderResolver(shardId)
	if err != nil {
		log.Error("getShardHeaderResolver",
			"error", err.Error(),
		)
		return
	}

	err = headerResolver.RequestDataFromHash(hash)
	if err != nil {
		log.Debug("RequestShardHeader", "error", err.Error())
		return
	}

	rrh.addRequestedItem(hash)
}

// RequestMetaHeader method asks for meta header from the connected peers
func (rrh *resolverRequestHandler) RequestMetaHeader(hash []byte) {
	if !rrh.testIfRequestIsNeeded(hash) {
		return
	}

	resolver, err := rrh.getMetaHeaderResolver()
	if err != nil {
		log.Error("RequestMetaHeader",
			"error", err.Error(),
		)
		return
	}

	err = resolver.RequestDataFromHash(hash)
	if err != nil {
		log.Debug("RequestDataFromHash", "error", err.Error())
		return
	}

	rrh.addRequestedItem(hash)
}

// RequestShardHeaderByNonce method asks for shard header from the connected peers by nonce
func (rrh *resolverRequestHandler) RequestShardHeaderByNonce(shardId uint32, nonce uint64) {
	key := []byte(fmt.Sprintf("%d-%d", shardId, nonce))
	if !rrh.testIfRequestIsNeeded(key) {
		return
	}

	headerResolver, err := rrh.getShardHeaderResolver(shardId)
	if err != nil {
		log.Error("RequestShardHeaderByNonce",
			"error", err.Error(),
		)
		return
	}

	err = headerResolver.RequestDataFromNonce(nonce)
	if err != nil {
		log.Debug("RequestShardHeaderByNonce", "error", err.Error())
		return
	}

	rrh.addRequestedItem(key)
}

// RequestMetaHeaderByNonce method asks for meta header from the connected peers by nonce
func (rrh *resolverRequestHandler) RequestMetaHeaderByNonce(nonce uint64) {
	key := []byte(fmt.Sprintf("%d-%d", sharding.MetachainShardId, nonce))
	if !rrh.testIfRequestIsNeeded(key) {
		return
	}

	headerResolver, err := rrh.getMetaHeaderResolver()
	if err != nil {
		log.Error("RequestMetaHeaderByNonce",
			"error", err.Error(),
		)
		return
	}

	err = headerResolver.RequestDataFromNonce(nonce)
	if err != nil {
		log.Debug("RequestMetaHeaderByNonce", "error", err.Error())
		return
	}

	rrh.addRequestedItem(key)
}

func (rrh *resolverRequestHandler) testIfRequestIsNeeded(key []byte) bool {
	rrh.sweepIfNeeded()

	if rrh.requestedItemsHandler.Has(string(key)) {
		log.Trace("item already requested",
			"key", key)
		return false
	}

	return true
}

func (rrh *resolverRequestHandler) addRequestedItem(key []byte) {
	err := rrh.requestedItemsHandler.Add(string(key))
	if err != nil {
		log.Debug("add requested item with error",
			"error", err.Error(),
			"key", key)
	}
}

func (rrh *resolverRequestHandler) getShardHeaderResolver(shardId uint32) (dataRetriever.HeaderResolver, error) {
	isMetachainNode := rrh.shardID == sharding.MetachainShardId
	shardIdMissmatch := rrh.shardID != shardId
	isRequestInvalid := !isMetachainNode && shardIdMissmatch
	if isRequestInvalid {
		return nil, dataRetriever.ErrBadRequest
	}

	//requests should be done on the topic shardBlocks_0_META so that is why we need to figure out
	//the cross shard id
	crossShardId := sharding.MetachainShardId
	if isMetachainNode {
		crossShardId = shardId
	}

	resolver, err := rrh.resolversFinder.CrossShardResolver(factory.ShardBlocksTopic, crossShardId)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, cross shard ID: %d",
			err, factory.ShardBlocksTopic, rrh.shardID, crossShardId)
		return nil, err
	}

	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, cross shard ID: %d, expected HeaderResolver",
			dataRetriever.ErrWrongTypeInContainer, factory.ShardBlocksTopic, rrh.shardID, crossShardId)
		return nil, err
	}

	return headerResolver, nil
}

func (rrh *resolverRequestHandler) getMetaHeaderResolver() (dataRetriever.HeaderResolver, error) {
	resolver, err := rrh.resolversFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d",
			err, factory.MetachainBlocksTopic, rrh.shardID)
		return nil, err
	}

	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, expected HeaderResolver",
			dataRetriever.ErrWrongTypeInContainer, factory.ShardBlocksTopic, rrh.shardID)
		return nil, err
	}

	return headerResolver, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *resolverRequestHandler) IsInterfaceNil() bool {
	return rrh == nil
}

func (rrh *resolverRequestHandler) getUnrequestedHashes(hashes [][]byte) [][]byte {
	unrequestedHashes := make([][]byte, 0)

	rrh.sweepIfNeeded()

	for _, hash := range hashes {
		if !rrh.requestedItemsHandler.Has(string(hash)) {
			unrequestedHashes = append(unrequestedHashes, hash)
			err := rrh.requestedItemsHandler.Add(string(hash))
			if err != nil {
				log.Trace("add requested item with error",
					"error", err.Error(),
					"key", hash)
			}
		}
	}

	return unrequestedHashes
}

func (rrh *resolverRequestHandler) sweepIfNeeded() {
	if time.Since(rrh.sweepTime) <= time.Second {
		return
	}

	rrh.sweepTime = time.Now()
	rrh.requestedItemsHandler.Sweep()
}
