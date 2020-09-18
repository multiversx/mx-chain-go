//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. miniblockMetadata.proto

package dblookupext

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/container"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

var log = logger.GetOrCreate("core/dblookupext")

const sizeOfDeduplicationCache = 1000

// HistoryRepositoryArguments is a structure that stores all components that are needed to a history processor
type HistoryRepositoryArguments struct {
	SelfShardID                 uint32
	MiniblocksMetadataStorer    storage.Storer
	MiniblockHashByTxHashStorer storage.Storer
	EpochByHashStorer           storage.Storer
	Marshalizer                 marshal.Marshalizer
	Hasher                      hashing.Hasher
}

type historyRepository struct {
	selfShardID                uint32
	miniblocksMetadataStorer   storage.Storer
	miniblockHashByTxHashIndex storage.Storer
	epochByHashIndex           *epochByHashIndex
	marshalizer                marshal.Marshalizer
	hasher                     hashing.Hasher

	// These maps temporarily hold notifications of "notarized at source or destination", to deal with unwanted concurrency effects
	// The unwanted concurrency effects could be accentuated by the fast db-replay-validate mechanism.
	pendingNotarizedAtSourceNotifications      *container.MutexMap
	pendingNotarizedAtDestinationNotifications *container.MutexMap
	pendingNotarizedAtBothNotifications        *container.MutexMap

	// This cache will hold hashes of already inserted miniblock metadata records, so that we avoid repeated "put()" operations,
	// that could mistakenly override the "patch()" operations performed when consuming notarization notifications.
	deduplicationCacheForInsertMiniblockMetadata storage.Cacher

	recordBlockMutex                 sync.Mutex
	consumePendingNotificationsMutex sync.Mutex
}

type notarizedNotification struct {
	metaNonce uint64
	metaHash  []byte
}

// NewHistoryRepository will create a new instance of HistoryRepository
func NewHistoryRepository(arguments HistoryRepositoryArguments) (*historyRepository, error) {
	if check.IfNil(arguments.MiniblocksMetadataStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.MiniblockHashByTxHashStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.EpochByHashStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.Marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return nil, core.ErrNilHasher
	}

	hashToEpochIndex := newHashToEpochIndex(arguments.EpochByHashStorer, arguments.Marshalizer)
	deduplicationCacheForInsertMiniblockMetadata, _ := lrucache.NewCache(sizeOfDeduplicationCache)

	return &historyRepository{
		selfShardID:                           arguments.SelfShardID,
		miniblocksMetadataStorer:              arguments.MiniblocksMetadataStorer,
		marshalizer:                           arguments.Marshalizer,
		hasher:                                arguments.Hasher,
		epochByHashIndex:                      hashToEpochIndex,
		miniblockHashByTxHashIndex:            arguments.MiniblockHashByTxHashStorer,
		pendingNotarizedAtSourceNotifications: container.NewMutexMap(),
		pendingNotarizedAtDestinationNotifications:   container.NewMutexMap(),
		pendingNotarizedAtBothNotifications:          container.NewMutexMap(),
		deduplicationCacheForInsertMiniblockMetadata: deduplicationCacheForInsertMiniblockMetadata,
	}, nil
}

// RecordBlock records a block
// This function is not called on a goroutine, but synchronously instead, right after committing a block
func (hr *historyRepository) RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error {
	hr.recordBlockMutex.Lock()
	defer hr.recordBlockMutex.Unlock()

	log.Debug("RecordBlock()", "nonce", blockHeader.GetNonce(), "blockHeaderHash", blockHeaderHash, "header type", fmt.Sprintf("%T", blockHeader))

	body, ok := blockBody.(*block.Body)
	if !ok {
		return errCannotCastToBlockBody
	}

	epoch := blockHeader.GetEpoch()

	err := hr.epochByHashIndex.saveEpochByHash(blockHeaderHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("block header", blockHeaderHash, err)
	}

	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		err = hr.recordMiniblock(blockHeaderHash, blockHeader, miniblock, epoch)
		if err != nil {
			continue
		}
	}

	return nil
}

func (hr *historyRepository) recordMiniblock(blockHeaderHash []byte, blockHeader data.HeaderHandler, miniblock *block.MiniBlock, epoch uint32) error {
	miniblockHash, err := hr.computeMiniblockHash(miniblock)
	if err != nil {
		return err
	}

	if hr.hasRecentlyInsertedMiniblockMetadata(miniblockHash, epoch) {
		return nil
	}

	err = hr.epochByHashIndex.saveEpochByHash(miniblockHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("miniblock", miniblockHash, err)
	}

	miniblockMetadata := &MiniblockMetadata{
		Type:               int32(miniblock.Type),
		Epoch:              epoch,
		HeaderHash:         blockHeaderHash,
		MiniblockHash:      miniblockHash,
		Round:              blockHeader.GetRound(),
		HeaderNonce:        blockHeader.GetNonce(),
		SourceShardID:      miniblock.GetSenderShardID(),
		DestinationShardID: miniblock.GetReceiverShardID(),
	}

	err = hr.putMiniblockMetadata(miniblockHash, miniblockMetadata)
	if err != nil {
		return err
	}

	hr.markMiniblockMetadataAsRecentlyInserted(miniblockHash, epoch)

	for _, txHash := range miniblock.TxHashes {
		err := hr.miniblockHashByTxHashIndex.Put(txHash, miniblockHash)
		if err != nil {
			log.Warn("miniblockHashByTxHashIndex.Put()", "txHash", txHash, "err", err)
			continue
		}
	}

	return nil
}

func (hr *historyRepository) computeMiniblockHash(miniblock *block.MiniBlock) ([]byte, error) {
	return core.CalculateHash(hr.marshalizer, hr.hasher, miniblock)
}

func (hr *historyRepository) hasRecentlyInsertedMiniblockMetadata(miniblockHash []byte, epoch uint32) bool {
	key := hr.buildKeyOfDeduplicationCacheForInsertMiniblockMetadata(miniblockHash, epoch)
	return hr.deduplicationCacheForInsertMiniblockMetadata.Has(key)
}

// When building the key for the deduplication cache, we must take into account the epoch as well, in order to handle this case:
// - miniblock M added in a fork at the end of epoch E,
// - miniblock M re-added, on the canonical chain this time, in the next epoch E + 1.
// This way we do not mistakenly ignore to update the "epochByHashIndex".
func (hr *historyRepository) buildKeyOfDeduplicationCacheForInsertMiniblockMetadata(miniblockHash []byte, epoch uint32) []byte {
	return []byte(fmt.Sprintf("%d_%x", epoch, miniblockHash))
}

func (hr *historyRepository) markMiniblockMetadataAsRecentlyInserted(miniblockHash []byte, epoch uint32) {
	key := hr.buildKeyOfDeduplicationCacheForInsertMiniblockMetadata(miniblockHash, epoch)
	_ = hr.deduplicationCacheForInsertMiniblockMetadata.Put(key, nil, 0)
}

// GetMiniblockMetadataByTxHash will return a history transaction for the given hash from storage
func (hr *historyRepository) GetMiniblockMetadataByTxHash(hash []byte) (*MiniblockMetadata, error) {
	miniblockHash, err := hr.miniblockHashByTxHashIndex.Get(hash)
	if err != nil {
		return nil, err
	}

	return hr.getMiniblockMetadataByMiniblockHash(miniblockHash)
}

func (hr *historyRepository) putMiniblockMetadata(hash []byte, metadata *MiniblockMetadata) error {
	metadataBytes, err := hr.marshalizer.Marshal(metadata)
	if err != nil {
		return err
	}

	err = hr.miniblocksMetadataStorer.PutInEpoch(hash, metadataBytes, metadata.Epoch)
	if err != nil {
		return newErrCannotSaveMiniblockMetadata(hash, err)
	}

	return nil
}

func (hr *historyRepository) getMiniblockMetadataByMiniblockHash(hash []byte) (*MiniblockMetadata, error) {
	epoch, err := hr.epochByHashIndex.getEpochByHash(hash)
	if err != nil {
		return nil, err
	}

	metadataBytes, err := hr.miniblocksMetadataStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	metadata := &MiniblockMetadata{}
	err = hr.marshalizer.Unmarshal(metadata, metadataBytes)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetEpochByHash will return epoch for a given hash
// This works for Blocks, Miniblocks
// It doesn't work for transactions (not needed, there we have a static storer for "miniblockHashByTxHashIndex" as well)!
func (hr *historyRepository) GetEpochByHash(hash []byte) (uint32, error) {
	return hr.epochByHashIndex.getEpochByHash(hash)
}

// OnNotarizedBlocks notifies the history repository about notarized blocks
func (hr *historyRepository) OnNotarizedBlocks(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	for i, headerHandler := range headers {
		headerHash := headersHashes[i]

		log.Debug("onNotarizedBlocks():", "shardID", shardID, "nonce", headerHandler.GetNonce(), "headerHash", headerHash, "type", fmt.Sprintf("%T", headerHandler))

		metaBlock, isMetaBlock := headerHandler.(*block.MetaBlock)
		if isMetaBlock {
			for _, miniBlock := range metaBlock.MiniBlockHeaders {
				hr.onNotarizedMiniblock(headerHandler.GetNonce(), headerHash, headerHandler.GetShardID(), miniBlock)
			}

			for _, shardData := range metaBlock.ShardInfo {
				hr.onNotarizedInMetaBlock(headerHandler.GetNonce(), headerHash, &shardData)
			}
		} else {
			log.Error("onNotarizedBlocks(): unexpected type of header", "type", fmt.Sprintf("%T", headerHandler))
		}
	}

	hr.consumePendingNotificationsWithLock()
}

func (hr *historyRepository) onNotarizedInMetaBlock(metaBlockNonce uint64, metaBlockHash []byte, shardData *block.ShardData) {
	if metaBlockNonce < 1 {
		return
	}

	for _, miniblockHeader := range shardData.GetShardMiniBlockHeaders() {
		hr.onNotarizedMiniblock(metaBlockNonce, metaBlockHash, shardData.GetShardID(), miniblockHeader)
	}
}

func (hr *historyRepository) onNotarizedMiniblock(metaBlockNonce uint64, metaBlockHash []byte, shardOfContainingBlock uint32, miniblockHeader block.MiniBlockHeader) {
	miniblockHash := miniblockHeader.Hash
	isIntra := miniblockHeader.SenderShardID == miniblockHeader.ReceiverShardID
	isToMeta := miniblockHeader.ReceiverShardID == core.MetachainShardId
	isNotarizedAtSource := miniblockHeader.SenderShardID == shardOfContainingBlock
	isNotarizedAtDestination := miniblockHeader.ReceiverShardID == shardOfContainingBlock
	isNotarizedAtBoth := isIntra || isToMeta

	notFromMe := miniblockHeader.SenderShardID != hr.selfShardID
	notToMe := miniblockHeader.ReceiverShardID != hr.selfShardID
	isPeerMiniblock := miniblockHeader.Type == block.PeerBlock
	iDontCare := (notFromMe && notToMe) || isPeerMiniblock
	if iDontCare {
		return
	}

	log.Debug("onNotarizedMiniblock()",
		"metaBlockNonce", metaBlockNonce,
		"metaBlockHash", metaBlockHash,
		"shardOfContainingBlock", shardOfContainingBlock,
		"miniblock", miniblockHash,
		"direction", fmt.Sprintf("[%d -> %d]", miniblockHeader.SenderShardID, miniblockHeader.ReceiverShardID),
	)

	if isNotarizedAtBoth {
		hr.pendingNotarizedAtBothNotifications.Set(string(miniblockHash), &notarizedNotification{
			metaNonce: metaBlockNonce,
			metaHash:  metaBlockHash,
		})
	} else if isNotarizedAtSource {
		hr.pendingNotarizedAtSourceNotifications.Set(string(miniblockHash), &notarizedNotification{
			metaNonce: metaBlockNonce,
			metaHash:  metaBlockHash,
		})
	} else if isNotarizedAtDestination {
		hr.pendingNotarizedAtDestinationNotifications.Set(string(miniblockHash), &notarizedNotification{
			metaNonce: metaBlockNonce,
			metaHash:  metaBlockHash,
		})
	} else {
		log.Error("onNotarizedMiniblock(): unexpected condition, notification not understood")
	}
}

// Notifications are consumed within a critical section so that we don't have competing put() operations for the same miniblock metadata,
// which could have resulted in mistakenly overriding the "notarization (hyperblock) coordinates".
func (hr *historyRepository) consumePendingNotificationsWithLock() {
	hr.consumePendingNotificationsMutex.Lock()
	defer hr.consumePendingNotificationsMutex.Unlock()

	if hr.pendingNotarizedAtSourceNotifications.Len() == 0 &&
		hr.pendingNotarizedAtDestinationNotifications.Len() == 0 &&
		hr.pendingNotarizedAtBothNotifications.Len() == 0 {
		return
	}

	log.Debug("consumePendingNotificationsWithLock() begin",
		"len(source)", hr.pendingNotarizedAtSourceNotifications.Len(),
		"len(destination)", hr.pendingNotarizedAtDestinationNotifications.Len(),
		"len(both)", hr.pendingNotarizedAtBothNotifications.Len(),
	)

	hr.consumePendingNotificationsNoLock(hr.pendingNotarizedAtSourceNotifications, func(metadata *MiniblockMetadata, notification *notarizedNotification) {
		metadata.NotarizedAtSourceInMetaNonce = notification.metaNonce
		metadata.NotarizedAtSourceInMetaHash = notification.metaHash
	})

	hr.consumePendingNotificationsNoLock(hr.pendingNotarizedAtDestinationNotifications, func(metadata *MiniblockMetadata, notification *notarizedNotification) {
		metadata.NotarizedAtDestinationInMetaNonce = notification.metaNonce
		metadata.NotarizedAtDestinationInMetaHash = notification.metaHash
	})

	hr.consumePendingNotificationsNoLock(hr.pendingNotarizedAtBothNotifications, func(metadata *MiniblockMetadata, notification *notarizedNotification) {
		metadata.NotarizedAtSourceInMetaNonce = notification.metaNonce
		metadata.NotarizedAtSourceInMetaHash = notification.metaHash
		metadata.NotarizedAtDestinationInMetaNonce = notification.metaNonce
		metadata.NotarizedAtDestinationInMetaHash = notification.metaHash
	})

	log.Debug("consumePendingNotificationsWithLock() end",
		"len(source)", hr.pendingNotarizedAtSourceNotifications.Len(),
		"len(destination)", hr.pendingNotarizedAtDestinationNotifications.Len(),
		"len(both)", hr.pendingNotarizedAtBothNotifications.Len(),
	)
}

func (hr *historyRepository) consumePendingNotificationsNoLock(pendingMap *container.MutexMap, patchMetadataFunc func(*MiniblockMetadata, *notarizedNotification)) {
	for _, key := range pendingMap.Keys() {
		notification, ok := pendingMap.Get(key)
		if !ok {
			continue
		}

		keyTyped, ok := key.(string)
		if !ok {
			log.Error("consumePendingNotificationsNoLock(): bad key", "key", key)
			continue
		}

		notificationTyped, ok := notification.(*notarizedNotification)
		if !ok {
			log.Error("consumePendingNotificationsNoLock(): bad value", "value", fmt.Sprintf("%T", notification))
			continue
		}

		miniblockHash := []byte(keyTyped)
		metadata, err := hr.getMiniblockMetadataByMiniblockHash(miniblockHash)
		if err != nil {
			// Maybe not yet committed / saved in storer
			continue
		}

		patchMetadataFunc(metadata, notificationTyped)
		err = hr.putMiniblockMetadata(miniblockHash, metadata)
		if err != nil {
			log.Error("consumePendingNotificationsNoLock(): cannot put miniblock metadata", "miniblockHash", miniblockHash, "err", err)
			continue
		}

		pendingMap.Remove(key)
	}
}

// IsEnabled will always returns true
func (hr *historyRepository) IsEnabled() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (hr *historyRepository) IsInterfaceNil() bool {
	return hr == nil
}
