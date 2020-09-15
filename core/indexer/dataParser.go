package indexer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type dataParser struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

func (dp *dataParser) getSerializedElasticBlockAndHeaderHash(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	sizeTxs int,
) ([]byte, []byte, error) {
	headerBytes, err := dp.marshalizer.Marshal(header)
	if err != nil {
		log.Debug("indexer: marshal header", "error", err)
		return nil, nil, err
	}
	bodyBytes, err := dp.marshalizer.Marshal(body)
	if err != nil {
		log.Debug("indexer: marshal body", "error", err)
		return nil, nil, err
	}

	blockSizeInBytes := len(headerBytes) + len(bodyBytes)

	miniblocksHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(dp.marshalizer, dp.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)
		miniblocksHashes = append(miniblocksHashes, encodedMbHash)
	}

	leaderIndex := uint64(0)
	if len(signersIndexes) > 0 {
		leaderIndex = signersIndexes[0]
	}

	headerHash := dp.hasher.Compute(string(headerBytes))
	elasticBlock := Block{
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
	}

	serializedBlock, err := json.Marshal(elasticBlock)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal elastic header")
		return nil, nil, err
	}

	return serializedBlock, headerHash, nil
}

func (dp *dataParser) getMiniblocks(header data.HeaderHandler, body *block.Body) []*Miniblock {
	headerHash, err := core.CalculateHash(dp.marshalizer, dp.hasher, header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err.Error())
		return nil
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)

	miniblocks := make([]*Miniblock, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(dp.marshalizer, dp.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("indexer: internal error computing hash",
				"error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)

		mb := &Miniblock{
			Hash:            encodedMbHash,
			SenderShardID:   miniblock.SenderShardID,
			ReceiverShardID: miniblock.ReceiverShardID,
			Type:            miniblock.Type.String(),
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

func serializeRoundInfo(info workItems.RoundInfo) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d_%d", "_type" : "%s" } }%s`,
		info.ShardId, info.Index, "_doc", "\n"))

	serializedInfo, err := json.Marshal(info)
	if err != nil {
		log.Warn("indexer: could not serialize round info, will skip indexing this round info",
			"error", err.Error())
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
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
