//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. miniblockMetadata.proto

package dblookupext

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common/logging"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dblookupext")

// HistoryRepositoryArguments is a structure that stores all components that are needed to a history processor
type HistoryRepositoryArguments struct {
	SelfShardID                 uint32
	MiniblocksMetadataStorer    storage.Storer
	MiniblockHashByTxHashStorer storage.Storer
	BlockHashByRound            storage.Storer
	Uint64ByteSliceConverter    typeConverters.Uint64ByteSliceConverter
	EpochByHashStorer           storage.Storer
	EventsHashesByTxHashStorer  storage.Storer
	Marshalizer                 marshal.Marshalizer
	Hasher                      hashing.Hasher
	ESDTSuppliesHandler         SuppliesHandler
}

type historyRepository struct {
	selfShardID                uint32
	miniblockHashByTxHashIndex storage.Storer
	blockHashByRound           storage.Storer
	uint64ByteSliceConverter   typeConverters.Uint64ByteSliceConverter
	epochByHashIndex           *epochByHashIndex
	eventsHashesByTxHashIndex  *eventsHashesByTxHash
	miniblocksHandler          *miniblocksHandler
	marshalizer                marshal.Marshalizer
	hasher                     hashing.Hasher
	esdtSuppliesHandler        SuppliesHandler

	recordBlockMutex                 sync.Mutex
	consumePendingNotificationsMutex sync.Mutex
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
	if check.IfNil(arguments.EventsHashesByTxHashStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(arguments.ESDTSuppliesHandler) {
		return nil, errNilESDTSuppliesHandler
	}
	if check.IfNil(arguments.Uint64ByteSliceConverter) {
		return nil, process.ErrNilUint64Converter
	}

	hashToEpochIndex := newHashToEpochIndex(arguments.EpochByHashStorer, arguments.Marshalizer)
	eventsHashesToTxHashIndex := newEventsHashesByTxHash(arguments.EventsHashesByTxHashStorer, arguments.Marshalizer)
	mbHandler := &miniblocksHandler{
		marshaller:                       arguments.Marshalizer,
		hasher:                           arguments.Hasher,
		epochIndex:                       hashToEpochIndex,
		miniblockHashByTxHashIndexStorer: arguments.MiniblockHashByTxHashStorer,
		miniblocksMetadataStorer:         arguments.MiniblocksMetadataStorer,
	}

	return &historyRepository{
		selfShardID:                arguments.SelfShardID,
		blockHashByRound:           arguments.BlockHashByRound,
		marshalizer:                arguments.Marshalizer,
		hasher:                     arguments.Hasher,
		epochByHashIndex:           hashToEpochIndex,
		miniblocksHandler:          mbHandler,
		miniblockHashByTxHashIndex: arguments.MiniblockHashByTxHashStorer,
		eventsHashesByTxHashIndex:  eventsHashesToTxHashIndex,
		esdtSuppliesHandler:        arguments.ESDTSuppliesHandler,
		uint64ByteSliceConverter:   arguments.Uint64ByteSliceConverter,
	}, nil
}

// RecordBlock records a block
// This function is not called on a goroutine, but synchronously instead, right after committing a block
func (hr *historyRepository) RecordBlock(blockHeaderHash []byte,
	blockHeader data.HeaderHandler,
	blockBody data.BodyHandler,
	scrResultsFromPool map[string]data.TransactionHandler,
	receiptsFromPool map[string]data.TransactionHandler,
	createdIntraShardMiniBlocks []*block.MiniBlock,
	logs []*data.LogData) error {
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

		err = hr.miniblocksHandler.commitMiniblock(blockHeader, blockHeaderHash, miniblock)
		if err != nil {
			logging.LogErrAsErrorExceptAsDebugIfClosingError(log, err, "cannot record miniblock",
				"type", miniblock.Type, "error", err)
			continue
		}
	}

	for _, miniblock := range createdIntraShardMiniBlocks {
		err = hr.miniblocksHandler.commitMiniblock(blockHeader, blockHeaderHash, miniblock)
		if err != nil {
			logging.LogErrAsErrorExceptAsDebugIfClosingError(log, err, "cannot record in shard miniblock",
				"type", miniblock.Type, "error", err)
		}
	}

	err = hr.eventsHashesByTxHashIndex.saveResultsHashes(epoch, scrResultsFromPool, receiptsFromPool)
	if err != nil {
		return err
	}

	err = hr.esdtSuppliesHandler.ProcessLogs(blockHeader.GetNonce(), logs)
	if err != nil {
		return err
	}

	err = hr.putHashByRound(blockHeaderHash, blockHeader)
	if err != nil {
		return err
	}

	return nil
}

func (hr *historyRepository) putHashByRound(blockHeaderHash []byte, header data.HeaderHandler) error {
	roundToByteSlice := hr.uint64ByteSliceConverter.ToByteSlice(header.GetRound())
	return hr.blockHashByRound.Put(roundToByteSlice, blockHeaderHash)
}

// GetMiniblockMetadataByTxHash will return a history transaction for the given hash from storage
func (hr *historyRepository) GetMiniblockMetadataByTxHash(txHash []byte) (*MiniblockMetadata, error) {
	return hr.miniblocksHandler.getMiniblockMetadataByTxHash(txHash)
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

		log.Trace("onNotarizedBlocks():", "shardID", shardID, "nonce", headerHandler.GetNonce(), "headerHash", headerHash, "type", fmt.Sprintf("%T", headerHandler))

		metaBlock, isMetaBlock := headerHandler.(*block.MetaBlock)
		if isMetaBlock {
			for _, miniBlock := range metaBlock.MiniBlockHeaders {
				hr.onNotarizedMiniblock(headerHandler.GetNonce(), headerHash, headerHandler.GetShardID(), miniBlock, headerHash)
			}

			for _, shardData := range metaBlock.ShardInfo {
				shardDataCopy := shardData
				hr.onNotarizedInMetaBlock(headerHandler.GetNonce(), headerHash, &shardDataCopy)
			}
		} else {
			log.Error("onNotarizedBlocks(): unexpected type of header", "type", fmt.Sprintf("%T", headerHandler))
		}
	}
}

func (hr *historyRepository) onNotarizedInMetaBlock(metaBlockNonce uint64, metaBlockHash []byte, shardData *block.ShardData) {
	if metaBlockNonce < 1 {
		return
	}

	for _, miniblockHeader := range shardData.GetShardMiniBlockHeaders() {
		hr.onNotarizedMiniblock(metaBlockNonce, metaBlockHash, shardData.GetShardID(), miniblockHeader, shardData.HeaderHash)
	}
}

func (hr *historyRepository) onNotarizedMiniblock(
	metaBlockNonce uint64,
	metaBlockHash []byte,
	shardOfContainingBlock uint32,
	miniblockHeader block.MiniBlockHeader,
	headerHash []byte,
) {
	miniblockHash := miniblockHeader.Hash

	notFromMe := miniblockHeader.SenderShardID != hr.selfShardID
	notToMe := miniblockHeader.ReceiverShardID != hr.selfShardID
	isPeerMiniblock := miniblockHeader.Type == block.PeerBlock
	iDontCare := (notFromMe && notToMe) || isPeerMiniblock
	if iDontCare {
		return
	}

	updateHandler := hr.createUpdateHandler(metaBlockHash, metaBlockNonce, miniblockHeader, shardOfContainingBlock)
	err := hr.miniblocksHandler.updateMiniblockMetadataOnBlock(miniblockHash, headerHash, updateHandler)
	if err != nil {
		log.Warn("historyRepository.onNotarizedMiniblock",
			"error", err,
			"shardOfContainingBlock", shardOfContainingBlock,
			"miniblockHeader.SenderShardID", miniblockHeader.SenderShardID,
			"miniblockHeader.ReceiverShardID", miniblockHeader.ReceiverShardID,
		)
	}
}

func (hr *historyRepository) createUpdateHandler(
	metaBlockHash []byte,
	metaBlockNonce uint64,
	miniblockHeader block.MiniBlockHeader,
	shardOfContainingBlock uint32,
) func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {

	isIntra := miniblockHeader.SenderShardID == miniblockHeader.ReceiverShardID
	isToMeta := miniblockHeader.ReceiverShardID == core.MetachainShardId
	isNotarizedAtSource := miniblockHeader.SenderShardID == shardOfContainingBlock
	isNotarizedAtDestination := miniblockHeader.ReceiverShardID == shardOfContainingBlock
	isNotarizedAtBoth := isIntra || isToMeta

	// the if checking order is important. Start checking for both and only after that for sender/receiver or
	// otherwise this function can return a wrong handler
	if isNotarizedAtBoth {
		return func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			mbMetadataOnBlock.NotarizedAtSourceInMetaHash = metaBlockHash
			mbMetadataOnBlock.NotarizedAtSourceInMetaNonce = metaBlockNonce
			mbMetadataOnBlock.NotarizedAtDestinationInMetaHash = metaBlockHash
			mbMetadataOnBlock.NotarizedAtDestinationInMetaNonce = metaBlockNonce
		}
	}
	if isNotarizedAtSource {
		return func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			mbMetadataOnBlock.NotarizedAtSourceInMetaHash = metaBlockHash
			mbMetadataOnBlock.NotarizedAtSourceInMetaNonce = metaBlockNonce
		}
	}
	if isNotarizedAtDestination {
		return func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
			mbMetadataOnBlock.NotarizedAtDestinationInMetaHash = metaBlockHash
			mbMetadataOnBlock.NotarizedAtDestinationInMetaNonce = metaBlockNonce
		}
	}

	return func(mbMetadataOnBlock *MiniblockMetadataOnBlock) {
		log.Warn("historyRepository.onNotarizedMiniblock - updateHandler programming error",
			"isNotarizedAtBoth", isNotarizedAtBoth,
			"isNotarizedAtSource", isNotarizedAtSource,
			"isNotarizedAtDestination", isNotarizedAtDestination,
		)
	}
}

// GetResultsHashesByTxHash will return results hashes by transaction hash
func (hr *historyRepository) GetResultsHashesByTxHash(txHash []byte, epoch uint32) (*ResultsHashesByTxHash, error) {
	return hr.eventsHashesByTxHashIndex.getEventsHashesByTxHash(txHash, epoch)
}

// IsEnabled will always return true
func (hr *historyRepository) IsEnabled() bool {
	return true
}

// RevertBlock will return the modification for the current block header
func (hr *historyRepository) RevertBlock(blockHeader data.HeaderHandler, blockBody data.BodyHandler) error {
	body, ok := blockBody.(*block.Body)
	if !ok {
		return fmt.Errorf("%w in historyRepository.RevertBlock for provided blockBody", errWrongTypeAssertion)
	}

	err := hr.miniblocksHandler.blockReverted(blockHeader, body)
	if err != nil {
		return err
	}

	return hr.esdtSuppliesHandler.RevertChanges(blockHeader, blockBody)
}

// GetESDTSupply will return the supply from the storage for the given token
func (hr *historyRepository) GetESDTSupply(token string) (*esdtSupply.SupplyESDT, error) {
	return hr.esdtSuppliesHandler.GetESDTSupply(token)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hr *historyRepository) IsInterfaceNil() bool {
	return hr == nil
}
