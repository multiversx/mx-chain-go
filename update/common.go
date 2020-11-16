package update

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TxInfo defines the structure which hold the tx info
type TxInfo struct {
	MbHash []byte
	TxHash []byte
	Tx     data.TransactionHandler
}

// GetPendingMiniBlocks get all the pending miniBlocks from epoch start metaBlock and unFinished metaBlocks
func GetPendingMiniBlocks(
	epochStartMetaBlock *block.MetaBlock,
	unFinishedMetaBlocksMap map[string]*block.MetaBlock,
) ([]block.MiniBlockHeader, error) {

	if epochStartMetaBlock == nil {
		return nil, ErrNilEpochStartMetaBlock
	}
	if unFinishedMetaBlocksMap == nil {
		return nil, ErrNilUnFinishedMetaBlocksMap
	}

	pendingMiniBlocks := make([]block.MiniBlockHeader, 0)
	nonceToHashMap := createNonceToHashMap(unFinishedMetaBlocksMap)

	for _, shardData := range epochStartMetaBlock.EpochStart.LastFinalizedHeaders {
		computedPendingMiniBlocks, err := computePendingMiniBlocksFromUnFinishedMetaBlocks(
			shardData,
			unFinishedMetaBlocksMap,
			nonceToHashMap,
			epochStartMetaBlock.GetNonce(),
		)
		if err != nil {
			return nil, err
		}

		pendingMiniBlocks = append(pendingMiniBlocks, computedPendingMiniBlocks...)
	}

	return pendingMiniBlocks, nil
}

// createNonceToHashMap creates a map of nonce to hash from all the given metaBlocks
func createNonceToHashMap(unFinishedMetaBlocks map[string]*block.MetaBlock) map[uint64]string {
	nonceToHashMap := make(map[uint64]string, len(unFinishedMetaBlocks))
	for metaBlockHash, metaBlock := range unFinishedMetaBlocks {
		nonceToHashMap[metaBlock.GetNonce()] = metaBlockHash
	}

	return nonceToHashMap
}

// computePendingMiniBlocksFromUnFinishedMetaBlocks computes all the pending miniBlocks from unFinished metaBlocks
func computePendingMiniBlocksFromUnFinishedMetaBlocks(
	epochStartShardData block.EpochStartShardData,
	unFinishedMetaBlocks map[string]*block.MetaBlock,
	nonceToHashMap map[uint64]string,
	epochStartMetaBlockNonce uint64,
) ([]block.MiniBlockHeader, error) {
	pendingMiniBlocks := make([]block.MiniBlockHeader, 0)
	pendingMiniBlocks = append(pendingMiniBlocks, epochStartShardData.PendingMiniBlockHeaders...)

	firstPendingMetaBlock, ok := unFinishedMetaBlocks[string(epochStartShardData.FirstPendingMetaBlock)]
	if !ok {
		return nil, ErrWrongUnFinishedMetaHdrsMap
	}

	firstUnFinishedMetaBlockNonce := firstPendingMetaBlock.GetNonce()
	for nonce := firstUnFinishedMetaBlockNonce + 1; nonce <= epochStartMetaBlockNonce; nonce++ {
		metaBlockHash, exists := nonceToHashMap[nonce]
		if !exists {
			return nil, ErrWrongUnFinishedMetaHdrsMap
		}

		metaBlock, exists := unFinishedMetaBlocks[metaBlockHash]
		if !exists {
			return nil, ErrWrongUnFinishedMetaHdrsMap
		}

		pendingMiniBlocksFromMetaBlock := getAllMiniBlocksWithDst(metaBlock, epochStartShardData.ShardID)
		pendingMiniBlocks = append(pendingMiniBlocks, pendingMiniBlocksFromMetaBlock...)
	}

	return pendingMiniBlocks, nil
}

// getAllMiniBlocksWithDst returns all miniBlock headers with the given destination from the given metaBlock
func getAllMiniBlocksWithDst(metaBlock *block.MetaBlock, destShardID uint32) []block.MiniBlockHeader {
	mbHdrs := make([]block.MiniBlockHeader, 0)
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		if metaBlock.ShardInfo[i].ShardID == destShardID {
			continue
		}

		for _, mbHdr := range metaBlock.ShardInfo[i].ShardMiniBlockHeaders {
			if mbHdr.ReceiverShardID == destShardID && mbHdr.SenderShardID != destShardID {
				mbHdrs = append(mbHdrs, mbHdr)
			}
		}
	}

	for _, mbHdr := range metaBlock.MiniBlockHeaders {
		if mbHdr.ReceiverShardID == destShardID && mbHdr.SenderShardID != destShardID {
			mbHdrs = append(mbHdrs, mbHdr)
		}
	}

	return mbHdrs
}
