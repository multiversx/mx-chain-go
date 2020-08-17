//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. miniblockMetadata.proto

package fullHistory

import (
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

func (hp *historyProcessor) RegisterToBlockTracker(blockTracker BlockTracker) {
	if check.IfNil(blockTracker) {
		log.Error("RegisterToBlockTracker(): blockTracker is nil")
		return
	}

	blockTracker.RegisterCrossNotarizedHeadersHandler(hp.onNotarizedBlockHeaders)
	blockTracker.RegisterSelfNotarizedHeadersHandler(hp.onNotarizedBlockHeaders)
}

func (hp *historyProcessor) onNotarizedBlockHeaders(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if shardID != core.MetachainShardId {
		return
	}

	log.Trace("onNotarizedBlockHeaders()", "shardID", shardID, "len(headers)", len(headers))

	// for i := 0; i < len(headers); i++ {
	// 	header := headers[i]
	// 	headerHash := headersHashes[i]

	// 	metaBlock, ok := header.(*block.MetaBlock)
	// 	if !ok {
	// 		log.Error("onNotarizedBlockHeaders(): cannot convert to *block.Metablock")
	// 		return
	// 	}

	// 	// shardHeaderHashes := make([]string, len(blockHeader.ShardInfo))
	// 	// for idx := 0; idx < len(blockHeader.ShardInfo); idx++ {
	// 	// 	shardHeaderHashes[idx] = hex.EncodeToString(blockHeader.ShardInfo[idx].HeaderHash)
	// 	// }
	// }
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

		err = hp.saveMiniblockMetadata(blockHeaderHash, blockHeader, miniblock, epoch)
		if err != nil {
			continue
		}
	}

	return nil
}

func (hp *historyProcessor) saveMiniblockMetadata(blockHeaderHash []byte, blockHeader data.HeaderHandler, miniblock *block.MiniBlock, epoch uint32) error {
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

	miniblockMetadataBytes, err := hp.marshalizer.Marshal(miniblockMetadata)
	if err != nil {
		return err
	}

	err = hp.miniblocksMetadataStorer.Put(miniblockHash, miniblockMetadataBytes)
	if err != nil {
		return newErrCannotSaveMiniblockMetadata(miniblockHash, err)
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

	epoch, err := hp.epochByHashIndex.getEpochByHash(miniblockHash)
	if err != nil {
		return nil, err
	}

	metadataBytes, err := hp.miniblocksMetadataStorer.GetFromEpoch(miniblockHash, epoch)
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
// It doesn't work for transactions (not needed!
func (hp *historyProcessor) GetEpochByHash(hash []byte) (uint32, error) {
	return hp.epochByHashIndex.getEpochByHash(hash)
}

// IsEnabled will always returns true
func (hp *historyProcessor) IsEnabled() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (hp *historyProcessor) IsInterfaceNil() bool {
	return hp == nil
}
