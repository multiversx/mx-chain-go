package preprocess

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type miniBlocksCompaction struct {
	economicsFee     process.FeeHandler
	shardCoordinator sharding.Coordinator

	mapHashToTx             map[string]data.TransactionHandler
	mapMinSenderNonce       map[string]uint64
	mapUnallocatedTxsHashes map[string]struct{}
	mapMiniBlockGasConsumed map[int]uint64

	mutMiniBlocksCompaction sync.RWMutex
}

// NewMiniBlocksCompaction creates a new mini blocks compaction object
func NewMiniBlocksCompaction(
	economicsFee process.FeeHandler,
	shardCoordinator sharding.Coordinator,
) (*miniBlocksCompaction, error) {

	if economicsFee == nil || economicsFee.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	mbc := miniBlocksCompaction{
		economicsFee:     economicsFee,
		shardCoordinator: shardCoordinator,
	}

	mbc.mapHashToTx = make(map[string]data.TransactionHandler)
	mbc.mapMinSenderNonce = make(map[string]uint64)
	mbc.mapUnallocatedTxsHashes = make(map[string]struct{})
	mbc.mapMiniBlockGasConsumed = make(map[int]uint64)

	return &mbc, nil
}

// Compact method tries to compact the given mini blocks to have only one mini block per sender/received pair
func (mbc *miniBlocksCompaction) Compact(
	miniBlocks block.MiniBlockSlice,
	mapHashToTx map[string]data.TransactionHandler,
) block.MiniBlockSlice {

	mbc.mutMiniBlocksCompaction.Lock()
	defer mbc.mutMiniBlocksCompaction.Unlock()

	if len(miniBlocks) <= 1 {
		return miniBlocks
	}

	var err error

	mbc.mapHashToTx = mapHashToTx
	mbc.mapMiniBlockGasConsumed = make(map[int]uint64)

	for index, miniBlock := range miniBlocks {
		mbc.mapMiniBlockGasConsumed[index], err = mbc.computeGasConsumedByMiniBlock(miniBlock)
		if err != nil {
			log.Error(err.Error())
			return miniBlocks
		}
	}

	compactedMiniBlocks := make(block.MiniBlockSlice, 0)
	compactedMiniBlocks = append(compactedMiniBlocks, miniBlocks[0])

	for index, miniBlock := range miniBlocks {
		if index == 0 {
			continue
		}

		compactedMiniBlocks = mbc.merge(compactedMiniBlocks, miniBlock, index)
	}

	if len(miniBlocks) > len(compactedMiniBlocks) {
		log.Info(fmt.Sprintf("compacted %d miniblocks to %d miniblocks\n",
			len(miniBlocks), len(compactedMiniBlocks)))
	}

	return compactedMiniBlocks
}

func (mbc *miniBlocksCompaction) merge(
	mergedMiniBlocks block.MiniBlockSlice,
	miniBlock *block.MiniBlock,
	srcIndex int,
) block.MiniBlockSlice {

	for dstIndex, mergedMiniBlock := range mergedMiniBlocks {
		sameType := miniBlock.Type == mergedMiniBlock.Type
		sameSenderShard := miniBlock.SenderShardID == mergedMiniBlock.SenderShardID
		sameReceiverShard := miniBlock.ReceiverShardID == mergedMiniBlock.ReceiverShardID

		canMerge := sameSenderShard && sameReceiverShard && sameType
		if canMerge {
			isFullyMerged := mbc.computeMerge(mergedMiniBlock, dstIndex, miniBlock, srcIndex)
			if isFullyMerged {
				return mergedMiniBlocks
			}
		}
	}

	mergedMiniBlocks = append(mergedMiniBlocks, miniBlock)

	return mergedMiniBlocks
}

func (mbc *miniBlocksCompaction) computeGasConsumedByMiniBlock(miniBlock *block.MiniBlock) (uint64, error) {
	gasUsedInMiniBlock := uint64(0)
	for _, txHash := range miniBlock.TxHashes {
		txGasLimit, err := mbc.computeGasConsumedByTx(txHash)
		if err != nil {
			return 0, err
		}

		gasUsedInMiniBlock += txGasLimit
	}

	return gasUsedInMiniBlock, nil
}

func (mbc *miniBlocksCompaction) computeGasConsumedByTx(txHash []byte) (uint64, error) {
	tx, ok := mbc.mapHashToTx[string(txHash)]
	if !ok {
		return 0, process.ErrMissingTransaction
	}

	txGasLimit := mbc.economicsFee.ComputeGasLimit(tx)
	if isSmartContractAddress(tx.GetRecvAddress()) {
		txGasLimit = tx.GetGasLimit()
	}

	return txGasLimit, nil
}

// Expand method tries to expand the given mini blocks to their initial state before compaction
func (mbc *miniBlocksCompaction) Expand(
	miniBlocks block.MiniBlockSlice,
	mapHashToTx map[string]data.TransactionHandler,
) (block.MiniBlockSlice, error) {

	mbc.mutMiniBlocksCompaction.Lock()
	defer mbc.mutMiniBlocksCompaction.Unlock()

	mbc.mapHashToTx = mapHashToTx
	mbc.mapMinSenderNonce = make(map[string]uint64)
	mbc.mapUnallocatedTxsHashes = make(map[string]struct{})

	expandedMiniBlocks := make(block.MiniBlockSlice, 0)
	miniBlocksToExpand := make(block.MiniBlockSlice, 0)

	for _, miniBlock := range miniBlocks {
		if miniBlock.SenderShardID == mbc.shardCoordinator.SelfId() {
			miniBlocksToExpand = append(miniBlocksToExpand, miniBlock)
			continue
		}

		expandedMiniBlocks = append(expandedMiniBlocks, miniBlock)
	}

	if len(miniBlocksToExpand) > 0 {
		expandedMiniBlocksFromMe, err := mbc.expandMiniBlocks(miniBlocksToExpand)
		if err != nil {
			return nil, err
		}

		expandedMiniBlocks = append(expandedMiniBlocks, expandedMiniBlocksFromMe...)
	}

	if len(miniBlocks) < len(expandedMiniBlocks) {
		log.Info(fmt.Sprintf("expanded from %d miniblocks to %d miniblocks\n",
			len(miniBlocks), len(expandedMiniBlocks)))
	}

	return expandedMiniBlocks, nil
}

func (mbc *miniBlocksCompaction) expandMiniBlocks(miniBlocks block.MiniBlockSlice) (block.MiniBlockSlice, error) {
	for _, miniBlock := range miniBlocks {
		for _, txHash := range miniBlock.TxHashes {
			tx, ok := mbc.mapHashToTx[string(txHash)]
			if !ok {
				return nil, process.ErrMissingTransaction
			}

			nonce, ok := mbc.mapMinSenderNonce[string(tx.GetSndAddress())]
			if !ok || nonce > tx.GetNonce() {
				mbc.mapMinSenderNonce[string(tx.GetSndAddress())] = tx.GetNonce()
			}

			mbc.mapUnallocatedTxsHashes[string(txHash)] = struct{}{}
		}
	}

	expandedMiniBlocks := make(block.MiniBlockSlice, 0)

	for len(mbc.mapUnallocatedTxsHashes) > 0 {
		createdMiniBlocks, err := mbc.createExpandedMiniBlocks(miniBlocks)
		if err != nil {
			return nil, err
		}

		if len(createdMiniBlocks) == 0 {
			break
		}

		expandedMiniBlocks = append(expandedMiniBlocks, createdMiniBlocks...)
	}

	return expandedMiniBlocks, nil
}

func (mbc *miniBlocksCompaction) createExpandedMiniBlocks(
	miniBlocks block.MiniBlockSlice,
) (block.MiniBlockSlice, error) {

	expandedMiniBlocks := make(block.MiniBlockSlice, 0)

	for _, miniBlock := range miniBlocks {
		if len(mbc.mapUnallocatedTxsHashes) == 0 {
			break
		}

		miniBlockForShard, err := mbc.createMiniBlockForShard(miniBlock)
		if err != nil {
			return nil, err
		}

		if len(miniBlockForShard.TxHashes) > 0 {
			expandedMiniBlocks = append(expandedMiniBlocks, miniBlockForShard)
		}
	}

	return expandedMiniBlocks, nil
}

func (mbc *miniBlocksCompaction) createMiniBlockForShard(miniBlock *block.MiniBlock) (*block.MiniBlock, error) {
	miniBlockForShard := &block.MiniBlock{}
	miniBlockForShard.TxHashes = make([][]byte, 0)
	miniBlockForShard.ReceiverShardID = miniBlock.ReceiverShardID
	miniBlockForShard.SenderShardID = miniBlock.SenderShardID
	miniBlockForShard.Type = miniBlock.Type

	for _, txHash := range miniBlock.TxHashes {
		if len(mbc.mapUnallocatedTxsHashes) == 0 {
			break
		}

		_, ok := mbc.mapUnallocatedTxsHashes[string(txHash)]
		if !ok {
			continue
		}

		tx, ok := mbc.mapHashToTx[string(txHash)]
		if !ok {
			return nil, process.ErrMissingTransaction
		}

		nonce := mbc.mapMinSenderNonce[string(tx.GetSndAddress())]
		if tx.GetNonce() == nonce {
			mbc.mapMinSenderNonce[string(tx.GetSndAddress())] = nonce + 1
			miniBlockForShard.TxHashes = append(miniBlockForShard.TxHashes, txHash)
			delete(mbc.mapUnallocatedTxsHashes, string(txHash))
		}
	}

	return miniBlockForShard, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbc *miniBlocksCompaction) IsInterfaceNil() bool {
	if mbc == nil {
		return true
	}
	return false
}

func (mbc *miniBlocksCompaction) computeMerge(
	dstMiniBlock *block.MiniBlock,
	dstIndex int,
	srcMiniBlock *block.MiniBlock,
	srcIndex int,
) bool {

	canBeFullyMerged := mbc.mapMiniBlockGasConsumed[dstIndex]+mbc.mapMiniBlockGasConsumed[srcIndex] <= process.MaxGasLimitPerMiniBlock
	if canBeFullyMerged {
		dstMiniBlock.TxHashes = append(dstMiniBlock.TxHashes, srcMiniBlock.TxHashes...)
		mbc.mapMiniBlockGasConsumed[dstIndex] += mbc.mapMiniBlockGasConsumed[srcIndex]
		return true
	}

	txHashes := make([][]byte, 0)
	dstMiniBlockGasConsumed := mbc.mapMiniBlockGasConsumed[dstIndex]
	for _, txHash := range srcMiniBlock.TxHashes {
		txGasLimit, err := mbc.computeGasConsumedByTx(txHash)
		if err != nil {
			break
		}

		isDstMiniBlockGasLimitReached := dstMiniBlockGasConsumed+txGasLimit > process.MaxGasLimitPerMiniBlock
		if isDstMiniBlockGasLimitReached {
			break
		}

		txHashes = append(txHashes, txHash)
		dstMiniBlockGasConsumed += txGasLimit
	}

	if len(txHashes) > 0 {
		miniBlockGasTransfered := dstMiniBlockGasConsumed - mbc.mapMiniBlockGasConsumed[dstIndex]
		dstMiniBlock.TxHashes = append(dstMiniBlock.TxHashes, txHashes...)
		srcMiniBlock.TxHashes = mbc.removeTxHashesFromMiniBlock(srcMiniBlock, txHashes)
		mbc.mapMiniBlockGasConsumed[dstIndex] += miniBlockGasTransfered
		mbc.mapMiniBlockGasConsumed[srcIndex] -= miniBlockGasTransfered
	}

	return false
}

func (mbc *miniBlocksCompaction) removeTxHashesFromMiniBlock(miniBlock *block.MiniBlock, txHashes [][]byte) [][]byte {
	mapTxHashesToBeRemoved := make(map[string]struct{})
	for _, txHash := range txHashes {
		mapTxHashesToBeRemoved[string(txHash)] = struct{}{}
	}

	preservedTxHashes := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		_, ok := mapTxHashesToBeRemoved[string(txHash)]
		if ok {
			continue
		}

		preservedTxHashes = append(preservedTxHashes, txHash)
	}

	return preservedTxHashes
}
