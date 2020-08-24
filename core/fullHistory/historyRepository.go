//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. miniblockMetadata.proto

package fullHistory

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/container"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("core/fullHistory")

// HistoryRepositoryArguments is a structure that stores all components that are needed to a history processor
type HistoryRepositoryArguments struct {
	SelfShardID                 uint32
	MiniblocksMetadataStorer    storage.Storer
	MiniblockHashByTxHashStorer storage.Storer
	EpochByHashStorer           storage.Storer
	Marshalizer                 marshal.Marshalizer
	Hasher                      hashing.Hasher
}

type historyProcessor struct {
	selfShardID                uint32
	miniblocksMetadataStorer   storage.Storer
	miniblockHashByTxHashIndex storage.Storer
	epochByHashIndex           *epochByHashIndex
	marshalizer                marshal.Marshalizer
	hasher                     hashing.Hasher

	// This map temporarily holds notifications of "notarized at source", for destination shard
	notarizedAtSourceNotifications *container.MutexMap
}

type notarizedAtSourceNotification struct {
	metaNonce uint64
	metaHash  []byte
}

// NewHistoryRepository will create a new instance of HistoryRepository
func NewHistoryRepository(arguments HistoryRepositoryArguments) (*historyProcessor, error) {
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

	return &historyProcessor{
		selfShardID:                    arguments.SelfShardID,
		miniblocksMetadataStorer:       arguments.MiniblocksMetadataStorer,
		marshalizer:                    arguments.Marshalizer,
		hasher:                         arguments.Hasher,
		epochByHashIndex:               hashToEpochIndex,
		miniblockHashByTxHashIndex:     arguments.MiniblockHashByTxHashStorer,
		notarizedAtSourceNotifications: container.NewMutexMap(),
	}, nil
}

// RecordBlock records a block
func (hp *historyProcessor) RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return errCannotCastToBlockBody
	}

	epoch := blockHeader.GetEpoch()

	err := hp.epochByHashIndex.saveEpochByHash(blockHeaderHash, epoch)
	if err != nil {
		return newErrCannotSaveEpochByHash("block header", blockHeaderHash, err)
	}

	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		err = hp.recordMiniblock(blockHeaderHash, blockHeader, miniblock, epoch)
		if err != nil {
			continue
		}
	}

	return nil
}

func (hp *historyProcessor) recordMiniblock(blockHeaderHash []byte, blockHeader data.HeaderHandler, miniblock *block.MiniBlock, epoch uint32) error {
	miniblockHash, err := hp.computeMiniblockHash(miniblock)
	if err != nil {
		return err
	}

	err = hp.epochByHashIndex.saveEpochByHash(miniblockHash, epoch)
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
		Status:             []byte(hp.getMiniblockStatus(miniblock)),
	}

	// Here we need to use queued notifications
	notification, ok := hp.notarizedAtSourceNotifications.Get(string(miniblockHash))
	if ok {
		notificationTyped := notification.(*notarizedAtSourceNotification)
		miniblockMetadata.NotarizedAtSourceInMetaNonce = notificationTyped.metaNonce
		miniblockMetadata.NotarizedAtSourceInMetaHash = notificationTyped.metaHash

		hp.notarizedAtSourceNotifications.Remove(string(miniblockHash))
	}

	err = hp.putMiniblockMetadata(miniblockHash, miniblockMetadata)
	if err != nil {
		return err
	}

	for _, txHash := range miniblock.TxHashes {
		err := hp.miniblockHashByTxHashIndex.Put(txHash, miniblockHash)
		if err != nil {
			log.Warn("miniblockHashByTxHashIndex.putMiniblockByTx()", "txHash", txHash, "err", err)
			continue
		}
	}

	return nil
}

func (hp *historyProcessor) computeMiniblockHash(miniblock *block.MiniBlock) ([]byte, error) {
	return core.CalculateHash(hp.marshalizer, hp.hasher, miniblock)
}

func (hp *historyProcessor) getMiniblockStatus(miniblock *block.MiniBlock) core.TransactionStatus {
	if miniblock.Type == block.InvalidBlock {
		return core.TxStatusInvalid
	}
	if miniblock.ReceiverShardID == hp.selfShardID {
		return core.TxStatusExecuted
	}

	return core.TxStatusPartiallyExecuted
}

// GetMiniblockMetadataByTxHash will return a history transaction for the given hash from storage
func (hp *historyProcessor) GetMiniblockMetadataByTxHash(hash []byte) (*MiniblockMetadata, error) {
	miniblockHash, err := hp.miniblockHashByTxHashIndex.Get(hash)
	if err != nil {
		return nil, err
	}

	return hp.getMiniblockMetadataByMiniblockHash(miniblockHash)
}

func (hp *historyProcessor) putMiniblockMetadata(hash []byte, metadata *MiniblockMetadata) error {
	metadataBytes, err := hp.marshalizer.Marshal(metadata)
	if err != nil {
		return err
	}

	err = hp.miniblocksMetadataStorer.Put(hash, metadataBytes)
	if err != nil {
		return newErrCannotSaveMiniblockMetadata(hash, err)
	}

	return nil
}

func (hp *historyProcessor) getMiniblockMetadataByMiniblockHash(hash []byte) (*MiniblockMetadata, error) {
	epoch, err := hp.epochByHashIndex.getEpochByHash(hash)
	if err != nil {
		return nil, err
	}

	metadataBytes, err := hp.miniblocksMetadataStorer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	metadata := &MiniblockMetadata{}
	err = hp.marshalizer.Unmarshal(metadata, metadataBytes)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetEpochByHash will return epoch for a given hash
// This works for Blocks, Miniblocks.
// It doesn't work for transactions (not needed)!
func (hp *historyProcessor) GetEpochByHash(hash []byte) (uint32, error) {
	return hp.epochByHashIndex.getEpochByHash(hash)
}

// RegisterToBlockTracker registers the history repository to blockTracker events
func (hp *historyProcessor) RegisterToBlockTracker(blockTracker BlockTracker) {
	if check.IfNil(blockTracker) {
		log.Error("RegisterToBlockTracker(): blockTracker is nil")
		return
	}

	blockTracker.RegisterCrossNotarizedHeadersHandler(hp.onNotarizedBlocks)
	blockTracker.RegisterSelfNotarizedHeadersHandler(hp.onNotarizedBlocks)
}

func (hp *historyProcessor) onNotarizedBlocks(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	log.Trace("onNotarizedBlocks()", "shardID", shardID, "len(headers)", len(headers))

	if shardID != core.MetachainShardId {
		return
	}

	for i, header := range headers {
		headerHash := headersHashes[i]

		metaBlock, ok := header.(*block.MetaBlock)
		if !ok {
			return
		}

		for _, shardData := range metaBlock.ShardInfo {
			hp.onNotarizedBlock(metaBlock.GetNonce(), headerHash, shardData)
		}
	}
}

func (hp *historyProcessor) onNotarizedBlock(metaBlockNonce uint64, metaBlockHash []byte, blockHeader block.ShardData) {
	for _, miniblockHeader := range blockHeader.GetShardMiniBlockHeaders() {
		hp.onNotarizedMiniblock(metaBlockNonce, metaBlockHash, blockHeader, miniblockHeader)
	}
}

func (hp *historyProcessor) onNotarizedMiniblock(metaBlockNonce uint64, metaBlockHash []byte, blockHeader block.ShardData, miniblockHeader block.MiniBlockHeader) {
	miniblockHash := miniblockHeader.Hash
	isIntra := miniblockHeader.SenderShardID == miniblockHeader.ReceiverShardID
	notarizedAtSource := miniblockHeader.SenderShardID == blockHeader.ShardID

	iDontCare := miniblockHeader.SenderShardID != hp.selfShardID && miniblockHeader.ReceiverShardID != hp.selfShardID
	if iDontCare {
		return
	}

	metadata, err := hp.getMiniblockMetadataByMiniblockHash(miniblockHash)
	if err != nil {
		if notarizedAtSource {
			// At destination, we receive the notification about "notarizedAtSource" before committing the block (at destination):
			// a) @source: source block is committed
			// b) @metachain: source block is notarized
			// c) @source & @destination: notified about b) 	<< we are here, under this condition
			// d) @destination: destination block is committed
			// e) @metachain: destination block is notarized
			// f) @source & @destination: notified about e)

			// Therefore, we should hold on to the notification at b) and use it at d)
			hp.notarizedAtSourceNotifications.Set(string(miniblockHash), &notarizedAtSourceNotification{
				metaNonce: metaBlockNonce,
				metaHash:  metaBlockHash,
			})
		} else {
			log.Warn("onNotarizedMiniblock() unexpected: cannot get miniblock metadata", "miniblock", miniblockHash, "err", err)
		}

		return
	}

	if isIntra {
		metadata.NotarizedAtSourceInMetaNonce = metaBlockNonce
		metadata.NotarizedAtSourceInMetaHash = metaBlockHash
		metadata.NotarizedAtDestinationInMetaNonce = metaBlockNonce
		metadata.NotarizedAtDestinationInMetaHash = metaBlockHash

		log.Trace("onNotarizedMiniblock() intra",
			"miniblock", miniblockHash,
			"direction", fmt.Sprintf("[%d -> %d]", metadata.SourceShardID, metadata.DestinationShardID),
			"meta nonce", metaBlockNonce,
		)
	} else {
		// Is cross-shard miniblock
		if notarizedAtSource {
			metadata.NotarizedAtSourceInMetaNonce = metaBlockNonce
			metadata.NotarizedAtSourceInMetaHash = metaBlockHash

			log.Trace("onNotarizedMiniblock() cross at source",
				"miniblock", miniblockHash,
				"direction", fmt.Sprintf("[%d -> %d]", metadata.SourceShardID, metadata.DestinationShardID),
				"meta nonce", metaBlockNonce,
			)
		} else {
			// Cross-shard, notarized at destination
			metadata.NotarizedAtDestinationInMetaNonce = metaBlockNonce
			metadata.NotarizedAtDestinationInMetaHash = metaBlockHash

			log.Trace("onNotarizedMiniblock() cross at destination",
				"miniblock", miniblockHash,
				"direction", fmt.Sprintf("[%d -> %d]", metadata.SourceShardID, metadata.DestinationShardID),
				"meta nonce", metaBlockNonce,
			)
		}
	}

	err = hp.putMiniblockMetadata(miniblockHash, metadata)
	if err != nil {
		log.Warn("onNotarizedMiniblock(): cannot update miniblock metadata", "miniblockHash", miniblockHash, "err", err)
		return
	}
}

// IsEnabled will always returns true
func (hp *historyProcessor) IsEnabled() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (hp *historyProcessor) IsInterfaceNil() bool {
	return hp == nil
}
