package blockAPI

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
)

// For each miniblock with processing type "processed" (miniblock "scheduled" in a previous block),
//  - for each transaction in the miniblock
//  - - recover the SCRs and the receipts (if any) and return them in an artificial miniblock (resembling the original, not persisted miniblock)
func (sbp *shardAPIBlockProcessor) recoverArtificialIntrashardMiniblocksHoldingContractResultsOfPreviouslyScheduledMiniblocks(
	header data.HeaderHandler,
	options api.BlockQueryOptions,
) ([]*api.MiniBlock, error) {
	miniblocks := make([]*api.MiniBlock, 0)

	for _, miniblockHeader := range header.GetMiniBlockHeaderHandlers() {
		// Only handle "processed" miniblocks (previously "scheduled")
		if miniblockHeader.GetProcessingType() != int32(block.Processed) {
			continue
		}
		// Only handle intrashard miniblocks (which aren't persisted)
		if miniblockHeader.GetSenderShardID() != miniblockHeader.GetReceiverShardID() {
			continue
		}

		miniblockOfSCRs, miniblockOfReceipts, err := sbp.recoverArtificialMiniblocksHoldingResults(
			header,
			miniblockHeader,
			options,
		)
		if err != nil {
			return nil, err
		}

		if len(miniblockOfSCRs.Transactions) > 0 {
			miniblocks = append(miniblocks, miniblockOfSCRs)
		}

		if len(miniblockOfReceipts.Receipts) > 0 {
			miniblocks = append(miniblocks, miniblockOfReceipts)
		}
	}

	return miniblocks, nil
}

func (sbp *shardAPIBlockProcessor) recoverArtificialMiniblocksHoldingResults(
	header data.HeaderHandler,
	miniblockHeader data.MiniBlockHeaderHandler,
	options api.BlockQueryOptions,
) (*api.MiniBlock, *api.MiniBlock, error) {
	epoch := header.GetEpoch()

	miniblock, err := sbp.getMiniblockByHash(miniblockHeader.GetHash(), epoch)
	if err != nil {
		return nil, nil, err
	}

	resultsByTx, err := sbp.historyRepo.GetResultsHashesByTxsHashes(miniblock.TxHashes, epoch)
	if err != nil {
		return nil, nil, err
	}

	scrsHashes, receiptsHashes := groupResultsIntoSCRsAndReceipts(resultsByTx)

	artificialMiniblockOfSCRs := &block.MiniBlock{
		TxHashes: scrsHashes,
		Type:     block.SmartContractResultBlock,
		// Set on purpose, to not be mistaken as "0" (actual shard not recoverable without extra grouping, not necessary)
		SenderShardID:   math.MaxUint16,
		ReceiverShardID: math.MaxUint16 + 1,
	}

	artificialMiniblockOfReceipts := &block.MiniBlock{
		TxHashes: receiptsHashes,
		Type:     block.ReceiptBlock,
		// Set on purpose, to not be mistaken as "0" (actual shard not recoverable without extra grouping, not necessary)
		SenderShardID:   math.MaxUint16,
		ReceiverShardID: math.MaxUint16 + 1,
	}

	artificialApiMiniblockOfSCRs := &api.MiniBlock{
		Hash: fmt.Sprintf("SCRs of %s", hex.EncodeToString(miniblockHeader.GetHash())),
		Type: block.SmartContractResultBlock.String(),
	}

	artificialApiMiniblockOfReceipts := &api.MiniBlock{
		Hash: fmt.Sprintf("Receipts of %s", hex.EncodeToString(miniblockHeader.GetHash())),
		Type: block.ReceiptBlock.String(),
	}

	err = sbp.getAndAttachTxsToMbByEpoch([]byte{}, artificialMiniblockOfSCRs, epoch, artificialApiMiniblockOfSCRs, options)
	if err != nil {
		return nil, nil, err
	}

	err = sbp.getAndAttachTxsToMbByEpoch([]byte{}, artificialMiniblockOfReceipts, epoch, artificialApiMiniblockOfReceipts, options)
	if err != nil {
		return nil, nil, err
	}

	return artificialApiMiniblockOfSCRs, artificialApiMiniblockOfReceipts, nil
}

func groupResultsIntoSCRsAndReceipts(results []*dblookupext.ResultsHashesByTxHashPair) ([][]byte, [][]byte) {
	scrsHashes := make([][]byte, 0)
	receiptsHashes := make([][]byte, 0)

	for _, item := range results {
		for _, innerItem := range item.ScResultsHashesAndEpoch {
			scrsHashes = append(scrsHashes, innerItem.ScResultsHashes...)
		}

		receiptsHashes = append(receiptsHashes, item.ReceiptsHash)
	}

	return scrsHashes, receiptsHashes
}
