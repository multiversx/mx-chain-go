package block

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("indexer/process/block")

type blockProcessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewBlockProcessor will create a new instance of block processor
func NewBlockProcessor(hasher hashing.Hasher, msarshalizer marshal.Marshalizer) *blockProcessor {
	return &blockProcessor{
		hasher:      hasher,
		marshalizer: msarshalizer,
	}
}

// PrepareBlockForDB will prepare a database block and serialize if for database
func (bp *blockProcessor) PrepareBlockForDB(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	sizeTxs int,
) (*types.Block, error) {
	blockSizeInBytes, headerHash, err := bp.computeBlockSizeAndHeaderHash(header, body)
	if err != nil {
		return nil, err
	}

	miniblocksHashes := bp.getEncodedMBSHashes(body)
	leaderIndex := bp.getLeaderIndex(signersIndexes)

	elasticBlock := &types.Block{
		Nonce:                 header.GetNonce(),
		Round:                 header.GetRound(),
		Epoch:                 header.GetEpoch(),
		ShardID:               header.GetShardID(),
		Hash:                  hex.EncodeToString(headerHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: notarizedHeadersHashes,
		Proposer:              leaderIndex,
		Validators:            signersIndexes,
		PubKeyBitmap:          hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             time.Duration(header.GetTimeStamp()),
		TxCount:               header.GetTxCount(),
		StateRootHash:         hex.EncodeToString(header.GetRootHash()),
		PrevHash:              hex.EncodeToString(header.GetPrevHash()),
		SearchOrder:           computeBlockSearchOrder(header),
		AccumulatedFees:       header.GetAccumulatedFees().String(),
		DeveloperFees:         header.GetDeveloperFees().String(),
		EpochStartBlock:       header.IsStartOfEpochBlock(),
	}

	return elasticBlock, nil
}

func (bp *blockProcessor) getEncodedMBSHashes(body *block.Body) []string {
	miniblocksHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(bp.marshalizer, bp.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)
		miniblocksHashes = append(miniblocksHashes, encodedMbHash)
	}

	return miniblocksHashes
}

func (bp *blockProcessor) computeBlockSizeAndHeaderHash(header data.HeaderHandler, body *block.Body) (int, []byte, error) {
	headerBytes, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return 0, nil, err
	}
	bodyBytes, err := bp.marshalizer.Marshal(body)
	if err != nil {
		return 0, nil, err
	}

	blockSize := len(headerBytes) + len(bodyBytes)

	headerHash := bp.hasher.Compute(string(headerBytes))

	return blockSize, headerHash, nil
}

func (bp *blockProcessor) getLeaderIndex(signersIndexes []uint64) uint64 {
	if len(signersIndexes) > 0 {
		return signersIndexes[0]
	}

	return 0
}

func computeBlockSearchOrder(header data.HeaderHandler) uint64 {
	shardIdentifier := createShardIdentifier(header.GetShardID())
	stringOrder := fmt.Sprintf("1%02d%d", shardIdentifier, header.GetNonce())

	order, err := strconv.ParseUint(stringOrder, 10, 64)
	if err != nil {
		log.Debug("elasticsearchDatabase.computeBlockSearchOrder",
			"could not set uint32 search order", err.Error())
		return 0
	}

	return order
}

func createShardIdentifier(shardID uint32) uint32 {
	shardIdentifier := shardID + 2
	if shardID == core.MetachainShardId {
		shardIdentifier = 1
	}

	return shardIdentifier
}

// ComputeHeaderHash will compute hash of a provided header
func (bp *blockProcessor) ComputeHeaderHash(header data.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(bp.marshalizer, bp.hasher, header)
}
