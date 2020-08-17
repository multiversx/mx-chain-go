//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. miniblockMetadata.proto

package fullHistory

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
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
	miniblockHashByTxHashIndex *miniblockHashByTxHashIndex
	epochByHashIndex           *epochByHashIndex
	marshalizer                marshal.Marshalizer
	hasher                     hashing.Hasher
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
	miniblockHashByTxHashIndex := newMiniblockHashByTxHashIndex(arguments.MiniblockHashByTxHashStorer)

	return &historyProcessor{
		selfShardID:                arguments.SelfShardID,
		miniblocksMetadataStorer:   arguments.MiniblocksMetadataStorer,
		marshalizer:                arguments.Marshalizer,
		hasher:                     arguments.Hasher,
		epochByHashIndex:           hashToEpochIndex,
		miniblockHashByTxHashIndex: miniblockHashByTxHashIndex,
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
		Epoch:              epoch,
		HeaderHash:         blockHeaderHash,
		MiniblockHash:      miniblockHash,
		Round:              blockHeader.GetRound(),
		HeaderNonce:        blockHeader.GetNonce(),
		SourceShardID:      miniblock.GetSenderShardID(),
		DestinationShardID: miniblock.GetReceiverShardID(),
		Status:             []byte(hp.getMiniblockStatus(miniblock)),
	}

	err = hp.putMiniblockMetadata(miniblockHash, miniblockMetadata)
	if err != nil {
		return err
	}

	for _, txHash := range miniblock.TxHashes {
		err := hp.miniblockHashByTxHashIndex.putMiniblockByTx(txHash, miniblockHash)
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
	miniblockHash, err := hp.miniblockHashByTxHashIndex.getMiniblockByTx(hash)
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

	for i := 0; i < len(headers); i++ {
		header := headers[i]
		headerHash := headersHashes[i]

		metaBlock, ok := header.(*block.MetaBlock)
		if !ok {
			// Question for review: why, on every round, there's a header of type = *block.Header (even if shardID = metachain)?
			log.Debug("onNotarizedBlocks(): cannot convert to *block.Metablock", "type", fmt.Sprintf("%T", header))
			return
		}

		log.Trace("onNotarizedBlocks()", "index", i, "len(ShardInfo)", len(metaBlock.ShardInfo))
		for _, shardData := range metaBlock.ShardInfo {
			hp.onNotarizedBlock(metaBlock.GetNonce(), headerHash, shardData)
		}
	}
}

func (hp *historyProcessor) onNotarizedBlock(metaBlockNonce uint64, metaBlockHash []byte, blockHeader block.ShardData) {
	log.Trace("onNotarizedBlock()",
		"metaBlockNonce", metaBlockNonce,
		"nonce", blockHeader.Nonce,
		"shard", blockHeader.ShardID,
		"len(miniblocks)", len(blockHeader.ShardMiniBlockHeaders),
	)

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
		log.Trace("onNotarizedMiniblock(): skipping",
			"miniblockHash", miniblockHash,
			"sender shard", miniblockHeader.SenderShardID,
			"receiver shard", miniblockHeader.ReceiverShardID,
		)
		return
	}

	metadata, err := hp.getMiniblockMetadataByMiniblockHash(miniblockHash)
	if err != nil {
		// Question for review: is there a way to be notified about the "source notarization", but at destination (cross-shard miniblock)?
		log.Debug("onNotarizedMiniblock(): cannot get miniblock metadata (yet)", "miniblockHash", miniblockHash, "err", err)
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
