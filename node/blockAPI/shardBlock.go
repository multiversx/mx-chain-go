package blockAPI

import (
	"bytes"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type shardAPIBlockProcessor struct {
	*baseAPIBockProcessor
}

// NewShardApiBlockProcessor will create a new instance of shard api block processor
func NewShardApiBlockProcessor(arg *APIBlockProcessorArg) *shardAPIBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()
	return &shardAPIBlockProcessor{
		baseAPIBockProcessor: &baseAPIBockProcessor{
			hasDbLookupExtensions:    hasDbLookupExtensions,
			selfShardID:              arg.SelfShardID,
			store:                    arg.Store,
			marshalizer:              arg.Marshalizer,
			uint64ByteSliceConverter: arg.Uint64ByteSliceConverter,
			historyRepo:              arg.HistoryRepo,
			unmarshalTx:              arg.UnmarshalTx,
		},
	}
}

// GetBlockByNonce will return a shard APIBlock by nonce
func (sbp *shardAPIBlockProcessor) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(sbp.selfShardID)

	nonceToByteSlice := sbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := sbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := sbp.getFromStorer(dataRetriever.BlockHeaderUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return sbp.convertShardBlockBytesToAPIBlock(headerHash, blockBytes, withTxs)
}

// GetBlockByHash will return a shard APIBlock by hash
func (sbp *shardAPIBlockProcessor) GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error) {
	blockBytes, err := sbp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	blockAPI, err := sbp.convertShardBlockBytesToAPIBlock(hash, blockBytes, withTxs)
	if err != nil {
		return nil, err
	}

	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(sbp.selfShardID)

	return sbp.computeStatusAndPutInBlock(blockAPI, storerUnit)
}

func (sbp *shardAPIBlockProcessor) convertShardBlockBytesToAPIBlock(hash []byte, blockBytes []byte, withTxs bool) (*api.Block, error) {
	blockHeader := &block.Header{}
	err := sbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	headerEpoch := blockHeader.Epoch

	numOfTxs := uint32(0)
	miniblocks := make([]*api.MiniBlock, 0)
	for _, mb := range blockHeader.MiniBlockHeaders {
		if mb.Type == block.PeerBlock {
			continue
		}

		numOfTxs += mb.TxCount

		miniblockAPI := &api.MiniBlock{
			Hash:             hex.EncodeToString(mb.Hash),
			Type:             mb.Type.String(),
			SourceShard:      mb.SenderShardID,
			DestinationShard: mb.ReceiverShardID,
		}
		if withTxs {
			miniBlockCopy := mb
			miniblockAPI.Transactions = sbp.getTxsByMb(&miniBlockCopy, headerEpoch)
		}

		miniblocks = append(miniblocks, miniblockAPI)
	}

	sbp.setStatusESDTTransferTransactionsCrossShard(miniblocks)

	return &api.Block{
		Nonce:           blockHeader.Nonce,
		Round:           blockHeader.Round,
		Epoch:           blockHeader.Epoch,
		Shard:           blockHeader.ShardID,
		Hash:            hex.EncodeToString(hash),
		PrevBlockHash:   hex.EncodeToString(blockHeader.PrevHash),
		NumTxs:          numOfTxs,
		MiniBlocks:      miniblocks,
		AccumulatedFees: blockHeader.AccumulatedFees.String(),
		DeveloperFees:   blockHeader.DeveloperFees.String(),
		Timestamp:       time.Duration(blockHeader.GetTimeStamp()),
		Status:          BlockStatusOnChain,
	}, nil
}

func (sbp *shardAPIBlockProcessor) setStatusESDTTransferTransactionsCrossShard(miniblocks []*api.MiniBlock) {
	for _, mb := range miniblocks {
		if mb.Type != block.TxBlock.String() {
			continue
		}

		// ignore miniblocks that are not cross shard destination me
		if mb.SourceShard == mb.DestinationShard ||
			mb.DestinationShard != sbp.selfShardID {
			continue
		}

		iterateMiniblockTxsForESDTTransfer(mb, miniblocks)
	}
}

func iterateMiniblockTxsForESDTTransfer(miniblock *api.MiniBlock, miniblocks []*api.MiniBlock) {
	for _, tx := range miniblock.Transactions {
		if !strings.HasPrefix(string(tx.Data), core.BuiltInFunctionESDTTransfer) {
			continue
		}

		// search for unsigned transactions
		for _, mb := range miniblocks {
			// search unsigned transaction from me to the source shard of the current transaction
			if !(mb.DestinationShard == tx.SourceShard && mb.SourceShard == tx.DestinationShard) {
				continue
			}

			tryToSetStatusOfESDTTransfer(tx, miniblock)
		}
	}
}

func tryToSetStatusOfESDTTransfer(tx *transaction.ApiTransactionResult, miniblock *api.MiniBlock) {
	for _, unsignedTx := range miniblock.Transactions {
		if unsignedTx.OriginalTransactionHash != tx.Hash {
			continue
		}

		if bytes.HasPrefix(unsignedTx.Data, tx.Data) && unsignedTx.Nonce == tx.Nonce {
			tx.Status = transaction.TxStatusFail
		}
	}
}
