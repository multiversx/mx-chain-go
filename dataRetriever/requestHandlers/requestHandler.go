package requestHandlers

import (
	"encoding/binary"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-logger-go"
)

var _ epochStart.RequestHandler = (*resolverRequestHandler)(nil)

var log = logger.GetOrCreate("dataretriever/requesthandlers")

const bytesInUint32 = 4
const timeToAccumulateTrieHashes = 100 * time.Millisecond
const uniqueTxSuffix = "tx"
const uniqueScrSuffix = "scr"
const uniqueRwdSuffix = "rwd"
const uniqueMiniblockSuffix = "mb"
const uniqueHeadersSuffix = "hdr"
const uniqueMetaHeadersSuffix = "mhdr"
const uniqueTrieNodesSuffix = "tn"
const uniqueValidatorInfoSuffix = "vi"

// TODO move the keys definitions that are whitelisted in core and use them in InterceptedData implementations, Identifiers() function

type resolverRequestHandler struct {
	mutEpoch              sync.RWMutex
	epoch                 uint32
	shardID               uint32
	maxTxsToRequest       int
	requestersFinder      dataRetriever.RequestersFinder
	requestedItemsHandler dataRetriever.RequestedItemsHandler
	whiteList             dataRetriever.WhiteListHandler
	sweepTime             time.Time
	requestInterval       time.Duration
	mutSweepTime          sync.Mutex

	trieHashesAccumulator map[string]struct{}
	lastTrieRequestTime   time.Time
	mutexTrieHashes       sync.Mutex
}

// NewResolverRequestHandler creates a requestHandler interface implementation with request functions
func NewResolverRequestHandler(
	finder dataRetriever.RequestersFinder,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	whiteList dataRetriever.WhiteListHandler,
	maxTxsToRequest int,
	shardID uint32,
	requestInterval time.Duration,
) (*resolverRequestHandler, error) {

	if check.IfNil(finder) {
		return nil, dataRetriever.ErrNilRequestersFinder
	}
	if check.IfNil(requestedItemsHandler) {
		return nil, dataRetriever.ErrNilRequestedItemsHandler
	}
	if maxTxsToRequest < 1 {
		return nil, dataRetriever.ErrInvalidMaxTxRequest
	}
	if check.IfNil(whiteList) {
		return nil, dataRetriever.ErrNilWhiteListHandler
	}
	if requestInterval < time.Millisecond {
		return nil, fmt.Errorf("%w:request interval is smaller than a millisecond", dataRetriever.ErrRequestIntervalTooSmall)
	}

	rrh := &resolverRequestHandler{
		requestersFinder:      finder,
		requestedItemsHandler: requestedItemsHandler,
		epoch:                 uint32(0), // will be updated after creation of the request handler
		shardID:               shardID,
		maxTxsToRequest:       maxTxsToRequest,
		whiteList:             whiteList,
		requestInterval:       requestInterval,
		trieHashesAccumulator: make(map[string]struct{}),
	}

	rrh.sweepTime = time.Now()

	return rrh, nil
}

// SetEpoch will update the current epoch so the request handler will make requests for this received epoch
func (rrh *resolverRequestHandler) SetEpoch(epoch uint32) {
	rrh.mutEpoch.Lock()
	if rrh.epoch != epoch {
		log.Debug("resolverRequestHandler.SetEpoch", "old epoch", rrh.epoch, "new epoch", epoch)
	}
	rrh.epoch = epoch
	rrh.mutEpoch.Unlock()
}

func (rrh *resolverRequestHandler) getEpoch() uint32 {
	rrh.mutEpoch.RLock()
	defer rrh.mutEpoch.RUnlock()

	return rrh.epoch
}

// RequestTransaction method asks for transactions from the connected peers
func (rrh *resolverRequestHandler) RequestTransaction(destShardID uint32, txHashes [][]byte) {
	rrh.requestByHashes(destShardID, txHashes, factory.TransactionTopic, uniqueTxSuffix)
}

func (rrh *resolverRequestHandler) requestByHashes(destShardID uint32, hashes [][]byte, topic string, abbreviatedTopic string) {
	suffix := fmt.Sprintf("%s_%d", abbreviatedTopic, destShardID)
	unrequestedHashes := rrh.getUnrequestedHashes(hashes, suffix)
	if len(unrequestedHashes) == 0 {
		return
	}
	log.Debug("requesting transactions from network",
		"topic", topic,
		"shard", destShardID,
		"num txs", len(unrequestedHashes),
	)
	requester, err := rrh.requestersFinder.CrossShardRequester(topic, destShardID)
	if err != nil {
		log.Error("requestByHashes.CrossShardRequester",
			"error", err.Error(),
			"topic", topic,
			"shard", destShardID,
		)
		return
	}

	txRequester, ok := requester.(HashSliceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating transaction requester")
		return
	}

	for _, txHash := range hashes {
		log.Trace("requestByHashes", "hash", txHash, "topic", topic,
			"shard", destShardID,
			"num txs", len(unrequestedHashes),
			"stack", string(debug.Stack()))
	}

	rrh.whiteList.Add(unrequestedHashes)

	go rrh.requestHashesWithDataSplit(unrequestedHashes, txRequester)

	rrh.addRequestedItems(unrequestedHashes, suffix)
}

func (rrh *resolverRequestHandler) requestHashesWithDataSplit(
	unrequestedHashes [][]byte,
	requester HashSliceRequester,
) {
	dataSplit := &partitioning.DataSplit{}
	sliceBatches, err := dataSplit.SplitDataInChunks(unrequestedHashes, rrh.maxTxsToRequest)
	if err != nil {
		log.Debug("requestByHashes.SplitDataInChunks",
			"error", err.Error(),
			"num txs", len(unrequestedHashes),
			"max txs to request", rrh.maxTxsToRequest,
		)
	}

	epoch := rrh.getEpoch()
	for _, batch := range sliceBatches {
		err = requester.RequestDataFromHashArray(batch, epoch)
		if err != nil {
			log.Debug("requestByHashes.RequestDataFromHashArray",
				"error", err.Error(),
				"epoch", epoch,
				"batch size", len(batch),
			)
		}
	}
}

func (rrh *resolverRequestHandler) requestReferenceWithChunkIndex(
	reference []byte,
	chunkIndex uint32,
	requester ChunkRequester,
) {
	err := requester.RequestDataFromReferenceAndChunk(reference, chunkIndex)
	if err != nil {
		log.Debug("requestByHashes.requestReferenceWithChunkIndex",
			"error", err.Error(),
			"reference", reference,
			"chunk index", chunkIndex,
		)
	}
}

// RequestUnsignedTransactions method asks for unsigned transactions from the connected peers
func (rrh *resolverRequestHandler) RequestUnsignedTransactions(destShardID uint32, scrHashes [][]byte) {
	rrh.requestByHashes(destShardID, scrHashes, factory.UnsignedTransactionTopic, uniqueScrSuffix)
}

// RequestRewardTransactions requests for reward transactions from the connected peers
func (rrh *resolverRequestHandler) RequestRewardTransactions(destShardID uint32, rewardTxHashes [][]byte) {
	rrh.requestByHashes(destShardID, rewardTxHashes, factory.RewardsTransactionTopic, uniqueRwdSuffix)
}

// RequestMiniBlock method asks for miniblock from the connected peers
func (rrh *resolverRequestHandler) RequestMiniBlock(destShardID uint32, miniblockHash []byte) {
	suffix := fmt.Sprintf("%s_%d", uniqueMiniblockSuffix, destShardID)
	if !rrh.testIfRequestIsNeeded(miniblockHash, suffix) {
		return
	}

	log.Debug("requesting miniblock from network",
		"topic", factory.MiniBlocksTopic,
		"shard", destShardID,
		"hash", miniblockHash,
	)

	requester, err := rrh.requestersFinder.CrossShardRequester(factory.MiniBlocksTopic, destShardID)
	if err != nil {
		log.Error("RequestMiniBlock.CrossShardRequester",
			"error", err.Error(),
			"topic", factory.MiniBlocksTopic,
			"shard", destShardID,
		)
		return
	}

	rrh.whiteList.Add([][]byte{miniblockHash})

	epoch := rrh.getEpoch()
	err = requester.RequestDataFromHash(miniblockHash, epoch)
	if err != nil {
		log.Debug("RequestMiniBlock.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", miniblockHash,
		)
		return
	}

	rrh.addRequestedItems([][]byte{miniblockHash}, suffix)
}

// RequestMiniBlocks method asks for miniblocks from the connected peers
func (rrh *resolverRequestHandler) RequestMiniBlocks(destShardID uint32, miniblocksHashes [][]byte) {
	suffix := fmt.Sprintf("%s_%d", uniqueMiniblockSuffix, destShardID)
	unrequestedHashes := rrh.getUnrequestedHashes(miniblocksHashes, suffix)
	if len(unrequestedHashes) == 0 {
		return
	}
	log.Debug("requesting miniblocks from network",
		"topic", factory.MiniBlocksTopic,
		"shard", destShardID,
		"num mbs", len(unrequestedHashes),
	)

	requester, err := rrh.requestersFinder.IntraShardRequester(factory.MiniBlocksTopic)
	if err != nil {
		log.Error("RequestMiniBlocks.CrossShardRequester",
			"error", err.Error(),
			"topic", factory.MiniBlocksTopic,
			"shard", destShardID,
		)
		return
	}

	miniBlocksRequester, ok := requester.(HashSliceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating miniblocks requester")
		return
	}

	rrh.whiteList.Add(unrequestedHashes)

	epoch := rrh.getEpoch()
	err = miniBlocksRequester.RequestDataFromHashArray(unrequestedHashes, epoch)
	if err != nil {
		log.Debug("RequestMiniBlocks.RequestDataFromHashArray",
			"error", err.Error(),
			"epoch", epoch,
			"num mbs", len(unrequestedHashes),
		)
		return
	}

	rrh.addRequestedItems(unrequestedHashes, suffix)
}

// RequestShardHeader method asks for shard header from the connected peers
func (rrh *resolverRequestHandler) RequestShardHeader(shardID uint32, hash []byte) {
	suffix := fmt.Sprintf("%s_%d", uniqueHeadersSuffix, shardID)
	if !rrh.testIfRequestIsNeeded(hash, suffix) {
		return
	}

	log.Debug("requesting shard header from network",
		"shard", shardID,
		"hash", hash,
	)

	headerRequester, err := rrh.getShardHeaderRequester(shardID)
	if err != nil {
		log.Error("RequestShardHeader.getShardHeaderRequester",
			"error", err.Error(),
			"shard", shardID,
		)
		return
	}

	rrh.whiteList.Add([][]byte{hash})

	epoch := rrh.getEpoch()
	err = headerRequester.RequestDataFromHash(hash, epoch)
	if err != nil {
		log.Debug("RequestShardHeader.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", hash,
		)
		return
	}

	rrh.addRequestedItems([][]byte{hash}, suffix)
}

// RequestMetaHeader method asks for meta header from the connected peers
func (rrh *resolverRequestHandler) RequestMetaHeader(hash []byte) {
	if !rrh.testIfRequestIsNeeded(hash, uniqueMetaHeadersSuffix) {
		return
	}

	log.Debug("requesting meta header from network",
		"hash", hash,
	)

	requester, err := rrh.getMetaHeaderRequester()
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

	rrh.whiteList.Add([][]byte{hash})

	epoch := rrh.getEpoch()
	err = headerRequester.RequestDataFromHash(hash, epoch)
	if err != nil {
		log.Debug("RequestMetaHeader.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", hash,
		)
		return
	}

	rrh.addRequestedItems([][]byte{hash}, uniqueMetaHeadersSuffix)
}

// RequestShardHeaderByNonce method asks for shard header from the connected peers by nonce
func (rrh *resolverRequestHandler) RequestShardHeaderByNonce(shardID uint32, nonce uint64) {
	suffix := fmt.Sprintf("%s_%d", uniqueHeadersSuffix, shardID)
	key := []byte(fmt.Sprintf("%d-%d", shardID, nonce))
	if !rrh.testIfRequestIsNeeded(key, suffix) {
		return
	}

	log.Debug("requesting shard header by nonce from network",
		"shard", shardID,
		"nonce", nonce,
	)

	requester, err := rrh.getShardHeaderRequester(shardID)
	if err != nil {
		log.Error("RequestShardHeaderByNonce.getShardHeaderRequester",
			"error", err.Error(),
			"shard", shardID,
		)
		return
	}

	headerRequester, ok := requester.(NonceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating header requester")
		return
	}

	rrh.whiteList.Add([][]byte{key})

	epoch := rrh.getEpoch()
	err = headerRequester.RequestDataFromNonce(nonce, epoch)
	if err != nil {
		log.Debug("RequestShardHeaderByNonce.RequestDataFromNonce",
			"error", err.Error(),
			"epoch", epoch,
			"nonce", nonce,
		)
		return
	}

	rrh.addRequestedItems([][]byte{key}, suffix)
}

// RequestTrieNodes method asks for trie nodes from the connected peers
func (rrh *resolverRequestHandler) RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string) {
	unrequestedHashes := rrh.getUnrequestedHashes(hashes, uniqueTrieNodesSuffix)
	if len(unrequestedHashes) == 0 {
		return
	}

	rrh.mutexTrieHashes.Lock()
	defer rrh.mutexTrieHashes.Unlock()

	for _, hash := range unrequestedHashes {
		rrh.trieHashesAccumulator[string(hash)] = struct{}{}
	}

	index := 0
	itemsToRequest := make([][]byte, len(rrh.trieHashesAccumulator))
	for hash := range rrh.trieHashesAccumulator {
		itemsToRequest[index] = []byte(hash)
		index++
	}

	rrh.whiteList.Add(itemsToRequest)

	elapsedTime := time.Since(rrh.lastTrieRequestTime)
	if elapsedTime < timeToAccumulateTrieHashes {
		return
	}

	log.Trace("requesting trie nodes from network",
		"topic", topic,
		"shard", destShardID,
		"num nodes", len(rrh.trieHashesAccumulator),
		"firstHash", unrequestedHashes[0],
		"added", len(hashes),
	)

	requester, err := rrh.requestersFinder.MetaCrossShardRequester(topic, destShardID)
	if err != nil {
		log.Error("requestersFinder.MetaCrossShardRequester",
			"error", err.Error(),
			"topic", topic,
			"shard", destShardID,
		)
		return
	}

	trieRequester, ok := requester.(HashSliceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating a trie nodes requester")
		return
	}

	rrh.logTrieHashesFromAccumulator()

	go rrh.requestHashesWithDataSplit(itemsToRequest, trieRequester)

	rrh.addRequestedItems(itemsToRequest, uniqueTrieNodesSuffix)
	rrh.lastTrieRequestTime = time.Now()
	rrh.trieHashesAccumulator = make(map[string]struct{})
}

// CreateTrieNodeIdentifier returns the requested trie node identifier that will be whitelisted
func (rrh *resolverRequestHandler) CreateTrieNodeIdentifier(requestHash []byte, chunkIndex uint32) []byte {
	chunkBuffer := make([]byte, bytesInUint32)
	binary.BigEndian.PutUint32(chunkBuffer, chunkIndex)

	return append(requestHash, chunkBuffer...)
}

// RequestTrieNode method asks for a trie node from the connected peers by the hash and the chunk index
func (rrh *resolverRequestHandler) RequestTrieNode(requestHash []byte, topic string, chunkIndex uint32) {
	identifier := rrh.CreateTrieNodeIdentifier(requestHash, chunkIndex)
	unrequestedHashes := rrh.getUnrequestedHashes([][]byte{identifier}, uniqueTrieNodesSuffix)
	if len(unrequestedHashes) == 0 {
		return
	}

	rrh.whiteList.Add(unrequestedHashes)
	rrh.whiteList.Add([][]byte{requestHash})

	log.Trace("requesting trie node from network",
		"topic", topic,
		"hash", requestHash,
		"chunk index", chunkIndex,
	)

	requester, err := rrh.requestersFinder.MetaChainRequester(topic)
	if err != nil {
		log.Error("requestersFinder.MetaChainRequester",
			"error", err.Error(),
			"topic", topic,
		)
		return
	}

	trieRequester, ok := requester.(ChunkRequester)
	if !ok {
		log.Warn("wrong assertion type when creating a trie chunk requester")
		return
	}

	go rrh.requestReferenceWithChunkIndex(requestHash, chunkIndex, trieRequester)

	rrh.addRequestedItems([][]byte{identifier}, uniqueTrieNodesSuffix)
}

func (rrh *resolverRequestHandler) logTrieHashesFromAccumulator() {
	if log.GetLevel() != logger.LogTrace {
		return
	}

	for txHash := range rrh.trieHashesAccumulator {
		log.Trace("logTrieHashesFromAccumulator", "hash", []byte(txHash))
	}
}

// RequestMetaHeaderByNonce method asks for meta header from the connected peers by nonce
func (rrh *resolverRequestHandler) RequestMetaHeaderByNonce(nonce uint64) {
	key := []byte(fmt.Sprintf("%d-%d", core.MetachainShardId, nonce))
	if !rrh.testIfRequestIsNeeded(key, uniqueMetaHeadersSuffix) {
		return
	}

	log.Debug("requesting meta header by nonce from network",
		"nonce", nonce,
	)

	headerRequester, err := rrh.getMetaHeaderRequester()
	if err != nil {
		log.Error("RequestMetaHeaderByNonce.getMetaHeaderRequester",
			"error", err.Error(),
		)
		return
	}

	rrh.whiteList.Add([][]byte{key})

	epoch := rrh.getEpoch()
	err = headerRequester.RequestDataFromNonce(nonce, epoch)
	if err != nil {
		log.Debug("RequestMetaHeaderByNonce.RequestDataFromNonce",
			"error", err.Error(),
			"epoch", epoch,
			"nonce", nonce,
		)
		return
	}

	rrh.addRequestedItems([][]byte{key}, uniqueMetaHeadersSuffix)
}

// RequestValidatorInfo asks for the validator info associated with a specific hash from connected peers
func (rrh *resolverRequestHandler) RequestValidatorInfo(hash []byte) {
	if !rrh.testIfRequestIsNeeded(hash, uniqueValidatorInfoSuffix) {
		return
	}

	log.Debug("requesting validator info messages from network",
		"topic", common.ValidatorInfoTopic,
		"hash", hash,
		"epoch", rrh.epoch,
	)

	requester, err := rrh.requestersFinder.MetaChainRequester(common.ValidatorInfoTopic)
	if err != nil {
		log.Error("RequestValidatorInfo.MetaChainRequester",
			"error", err.Error(),
			"topic", common.ValidatorInfoTopic,
			"hash", hash,
			"epoch", rrh.epoch,
		)
		return
	}

	rrh.whiteList.Add([][]byte{hash})

	err = requester.RequestDataFromHash(hash, rrh.epoch)
	if err != nil {
		log.Debug("RequestValidatorInfo.RequestDataFromHash",
			"error", err.Error(),
			"topic", common.ValidatorInfoTopic,
			"hash", hash,
			"epoch", rrh.epoch,
		)
		return
	}

	rrh.addRequestedItems([][]byte{hash}, uniqueValidatorInfoSuffix)
}

// RequestValidatorsInfo asks for the validators` info associated with the specified hashes from connected peers
func (rrh *resolverRequestHandler) RequestValidatorsInfo(hashes [][]byte) {
	unrequestedHashes := rrh.getUnrequestedHashes(hashes, uniqueValidatorInfoSuffix)
	if len(unrequestedHashes) == 0 {
		return
	}

	log.Debug("requesting validator info messages from network",
		"topic", common.ValidatorInfoTopic,
		"num hashes", len(unrequestedHashes),
		"epoch", rrh.epoch,
	)

	requester, err := rrh.requestersFinder.MetaChainRequester(common.ValidatorInfoTopic)
	if err != nil {
		log.Error("RequestValidatorInfo.MetaChainRequester",
			"error", err.Error(),
			"topic", common.ValidatorInfoTopic,
			"num hashes", len(unrequestedHashes),
			"epoch", rrh.epoch,
		)
		return
	}

	validatorInfoRequester, ok := requester.(HashSliceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating a validator info requester")
		return
	}

	rrh.whiteList.Add(unrequestedHashes)

	err = validatorInfoRequester.RequestDataFromHashArray(unrequestedHashes, rrh.epoch)
	if err != nil {
		log.Debug("RequestValidatorInfo.RequestDataFromHash",
			"error", err.Error(),
			"topic", common.ValidatorInfoTopic,
			"num hashes", len(unrequestedHashes),
			"epoch", rrh.epoch,
		)
		return
	}

	rrh.addRequestedItems(unrequestedHashes, uniqueValidatorInfoSuffix)
}

func (rrh *resolverRequestHandler) testIfRequestIsNeeded(key []byte, suffix string) bool {
	rrh.sweepIfNeeded()

	if rrh.requestedItemsHandler.Has(string(key) + suffix) {
		log.Trace("item already requested",
			"key", key)
		return false
	}

	return true
}

func (rrh *resolverRequestHandler) addRequestedItems(keys [][]byte, suffix string) {
	for _, key := range keys {
		err := rrh.requestedItemsHandler.Add(string(key) + suffix)
		if err != nil {
			log.Trace("addRequestedItems",
				"error", err.Error(),
				"key", key)
			continue
		}
	}
}

func (rrh *resolverRequestHandler) getShardHeaderRequester(shardID uint32) (dataRetriever.Requester, error) {
	isMetachainNode := rrh.shardID == core.MetachainShardId
	shardIdMissmatch := rrh.shardID != shardID
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

	//headerRequester, err := rrh.requestersFinder.CrossShardRequester(factory.ShardBlocksTopic, core.SovereignChainShardId)
	headerRequester, err := rrh.requestersFinder.CrossShardRequester(factory.ShardBlocksTopic, crossShardID)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, cross shard ID: %d",
			err, factory.ShardBlocksTopic, rrh.shardID, crossShardID)

		log.Warn("available requesters in container",
			"requesters", rrh.requestersFinder.RequesterKeys(),
		)
		return nil, err
	}

	return headerRequester, nil
}

func (rrh *resolverRequestHandler) getMetaHeaderRequester() (HeaderRequester, error) {
	requester, err := rrh.requestersFinder.MetaChainRequester(factory.MetachainBlocksTopic)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d",
			err, factory.MetachainBlocksTopic, rrh.shardID)
		return nil, err
	}

	headerRequester, ok := requester.(HeaderRequester)
	if !ok {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, expected HeaderRequester",
			dataRetriever.ErrWrongTypeInContainer, factory.ShardBlocksTopic, rrh.shardID)
		return nil, err
	}

	return headerRequester, nil
}

// RequestStartOfEpochMetaBlock method asks for the start of epoch metablock from the connected peers
func (rrh *resolverRequestHandler) RequestStartOfEpochMetaBlock(epoch uint32) {
	epochStartIdentifier := []byte(core.EpochStartIdentifier(epoch))
	if !rrh.testIfRequestIsNeeded(epochStartIdentifier, uniqueMetaHeadersSuffix) {
		return
	}

	baseTopic := factory.MetachainBlocksTopic
	log.Debug("requesting header by epoch",
		"topic", baseTopic,
		"epoch", epoch,
		"hash", epochStartIdentifier,
	)

	requester, err := rrh.requestersFinder.MetaChainRequester(baseTopic)
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

	rrh.whiteList.Add([][]byte{epochStartIdentifier})

	err = headerRequester.RequestDataFromEpoch(epochStartIdentifier)
	if err != nil {
		log.Debug("RequestStartOfEpochMetaBlock.RequestDataFromEpoch",
			"error", err.Error(),
			"epochStartIdentifier", epochStartIdentifier,
		)
		return
	}

	rrh.addRequestedItems([][]byte{epochStartIdentifier}, uniqueMetaHeadersSuffix)
}

// RequestInterval returns the request interval between sending the same request
func (rrh *resolverRequestHandler) RequestInterval() time.Duration {
	return rrh.requestInterval
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *resolverRequestHandler) IsInterfaceNil() bool {
	return rrh == nil
}

func (rrh *resolverRequestHandler) getUnrequestedHashes(hashes [][]byte, suffix string) [][]byte {
	unrequestedHashes := make([][]byte, 0)

	rrh.sweepIfNeeded()

	for _, hash := range hashes {
		if !rrh.requestedItemsHandler.Has(string(hash) + suffix) {
			unrequestedHashes = append(unrequestedHashes, hash)
		}
	}

	return unrequestedHashes
}

func (rrh *resolverRequestHandler) sweepIfNeeded() {
	rrh.mutSweepTime.Lock()
	defer rrh.mutSweepTime.Unlock()

	if time.Since(rrh.sweepTime) <= rrh.requestInterval {
		return
	}

	rrh.sweepTime = time.Now()
	rrh.requestedItemsHandler.Sweep()
}

// SetNumPeersToQuery will set the number of intra shard and cross shard number of peers to query
// for a given requester
func (rrh *resolverRequestHandler) SetNumPeersToQuery(key string, intra int, cross int) error {
	requester, err := rrh.requestersFinder.Get(key)
	if err != nil {
		return err
	}

	requester.SetNumPeersToQuery(intra, cross)
	return nil
}

// GetNumPeersToQuery will return the number of intra shard and cross shard number of peers to query
// for a given requester
func (rrh *resolverRequestHandler) GetNumPeersToQuery(key string) (int, int, error) {
	requester, err := rrh.requestersFinder.Get(key)
	if err != nil {
		return 0, 0, err
	}

	intra, cross := requester.NumPeersToQuery()
	return intra, cross, nil
}

// RequestPeerAuthenticationsByHashes asks for peer authentication messages from specific peers hashes
func (rrh *resolverRequestHandler) RequestPeerAuthenticationsByHashes(destShardID uint32, hashes [][]byte) {
	log.Debug("requesting peer authentication messages from network",
		"topic", common.PeerAuthenticationTopic,
		"shard", destShardID,
		"num hashes", len(hashes),
		"epoch", rrh.epoch,
	)

	requester, err := rrh.requestersFinder.MetaChainRequester(common.PeerAuthenticationTopic)
	if err != nil {
		log.Error("RequestPeerAuthenticationsByHashes.MetaChainRequester",
			"error", err.Error(),
			"topic", common.PeerAuthenticationTopic,
			"shard", destShardID,
			"epoch", rrh.epoch,
		)
		return
	}

	peerAuthRequester, ok := requester.(HashSliceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating peer authentication requester")
		return
	}

	err = peerAuthRequester.RequestDataFromHashArray(hashes, rrh.epoch)
	if err != nil {
		log.Debug("RequestPeerAuthenticationsByHashes.RequestDataFromHashArray",
			"error", err.Error(),
			"topic", common.PeerAuthenticationTopic,
			"shard", destShardID,
			"epoch", rrh.epoch,
		)
	}
}
