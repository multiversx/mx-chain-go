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
	hasher     hashing.Hasher
	marshalier marshal.Marshalizer
}

// NewMiniblocksProcessor will create a new instance of miniblocksProcessor
func NewMiniblocksProcessor(hasher hashing.Hasher, marshalier marshal.Marshalizer) *miniblocksProcessor {
	return &miniblocksProcessor{
		hasher:     hasher,
		marshalier: marshalier,
	}
}

// PrepareDBMiniblocks -
func (mp *miniblocksProcessor) PrepareDBMiniblocks(header data.HeaderHandler, body *block.Body) []*types.Miniblock {
	headerHash, err := mp.calculateHash(header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err.Error())
		return nil
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)

	miniblocks := make([]*types.Miniblock, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := mp.calculateHash(miniblock)
		if errComputeHash != nil {
			log.Warn("indexer: internal error computing hash",
				"error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)

		mb := &types.Miniblock{
			Hash:            encodedMbHash,
			SenderShardID:   miniblock.SenderShardID,
			ReceiverShardID: miniblock.ReceiverShardID,
			Type:            miniblock.Type.String(),
			Timestamp:       time.Duration(header.GetTimeStamp()),
		}

		if mb.SenderShardID == header.GetShardID() {
			mb.SenderBlockHash = encodedHeaderHash
		} else {
			mb.ReceiverBlockHash = encodedHeaderHash
		}

		if mb.SenderShardID == mb.ReceiverShardID {
			mb.ReceiverBlockHash = encodedHeaderHash
		}

		miniblocks = append(miniblocks, mb)
	}

	return miniblocks
}

func (mp *miniblocksProcessor) calculateHash(object interface{}) ([]byte, error) {
	return core.CalculateHash(mp.marshalier, mp.hasher, object)
}
