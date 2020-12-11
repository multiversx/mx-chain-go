package process

import (
	"encoding/hex"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/websockets/types"
)

var log = logger.GetOrCreate("websockets/process")

type blockProcessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

func newBlockProcessor(hasher hashing.Hasher, marshalizer marshal.Marshalizer) (*blockProcessor, error) {
	return &blockProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
	}, nil
}

func (bp *blockProcessor) prepareBlock(
	header data.HeaderHandler,
	signersIndexes []uint64,
	bodyHandler data.BodyHandler,
	notarizedHeadersHashes []string,
	txs map[string]data.TransactionHandler,
) (*types.Block, error) {
	headerBytes, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return nil, err
	}

	headerHash := bp.hasher.Compute(string(headerBytes))
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, fmt.Errorf("when trying body assertion, block hash %s, nonce %d",
			headerHash, header.GetNonce())
	}

	bodyBytes, err := bp.marshalizer.Marshal(body)
	if err != nil {
		return nil, err
	}

	blockSizeInBytes := len(headerBytes) + len(bodyBytes)

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

	leaderIndex := uint64(0)
	if len(signersIndexes) > 0 {
		leaderIndex = signersIndexes[0]
	}

	sizeTxs := drivers.ComputeSizeOfTxs(bp.marshalizer, txs)

	return &types.Block{
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
	}, nil
}
