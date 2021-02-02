package miniblocks

import (
	"encoding/hex"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("indexer/process/miniblocks")

type miniblocksProcessor struct {
	hasher      hashing.Hasher
	marshalier  marshal.Marshalizer
	selfShardID uint32
}

// NewMiniblocksProcessor will create a new instance of miniblocksProcessor
func NewMiniblocksProcessor(
	selfShardID uint32,
	hasher hashing.Hasher,
	marshalier marshal.Marshalizer,
) *miniblocksProcessor {
	return &miniblocksProcessor{
		hasher:      hasher,
		marshalier:  marshalier,
		selfShardID: selfShardID,
	}
}

// PrepareDBMiniblocks -
func (mp *miniblocksProcessor) PrepareDBMiniblocks(header data.HeaderHandler, body *block.Body) []*types.Miniblock {
	headerHash, err := mp.calculateHash(header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err.Error())
		return nil
	}

	dbMiniblocks := make([]*types.Miniblock, 0)
	for _, miniblock := range body.MiniBlocks {
		dbMiniblock, errPrepareMiniblock := mp.prepareMiniblockForDB(miniblock, header, headerHash)
		if errPrepareMiniblock != nil {
			log.Warn("miniblocksProcessor.PrepareDBMiniblocks cannot prepare miniblock", "error", errPrepareMiniblock)
		}

		dbMiniblocks = append(dbMiniblocks, dbMiniblock)
	}

	return dbMiniblocks
}
func (mp *miniblocksProcessor) prepareMiniblockForDB(
	miniblock *block.MiniBlock,
	header data.HeaderHandler,
	headerHash []byte,
) (*types.Miniblock, error) {
	mbHash, err := mp.calculateHash(miniblock)
	if err != nil {
		return nil, err
	}

	encodedMbHash := hex.EncodeToString(mbHash)

	dbMiniblock := &types.Miniblock{
		Hash:            encodedMbHash,
		SenderShardID:   miniblock.SenderShardID,
		ReceiverShardID: miniblock.ReceiverShardID,
		Type:            miniblock.Type.String(),
		Timestamp:       time.Duration(header.GetTimeStamp()),
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)
	if dbMiniblock.SenderShardID == header.GetShardID() {
		dbMiniblock.SenderBlockHash = encodedHeaderHash
	} else {
		dbMiniblock.ReceiverBlockHash = encodedHeaderHash
	}

	if dbMiniblock.SenderShardID == dbMiniblock.ReceiverShardID {
		dbMiniblock.ReceiverBlockHash = encodedHeaderHash
	}

	return dbMiniblock, nil
}

// GetMiniblocksHashesHexEncoded will compute miniblocks hashes hex encoded
func (mp *miniblocksProcessor) GetMiniblocksHashesHexEncoded(header data.HeaderHandler, body *block.Body) []string {
	if body == nil || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	encodedMiniblocksHashes := make([]string, 0)
	selfShardID := header.GetShardID()
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		isCrossShard := miniblock.ReceiverShardID != miniblock.SenderShardID
		if isDstMe && isCrossShard {
			continue
		}

		miniblockHash, err := core.CalculateHash(mp.marshalier, mp.hasher, miniblock)
		if err != nil {
			log.Debug("RemoveMiniblocks cannot calculate miniblock hash",
				"error", err.Error())
			continue
		}
		encodedMiniblocksHashes = append(encodedMiniblocksHashes, hex.EncodeToString(miniblockHash))
	}

	return encodedMiniblocksHashes
}

func (mp *miniblocksProcessor) calculateHash(object interface{}) ([]byte, error) {
	return core.CalculateHash(mp.marshalier, mp.hasher, object)
}
