package preprocess

import (
	"bytes"
	"errors"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

func (txs *transactions) createAndProcessMiniBlocksFromMeV2(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, error) {
	log.Debug("createAndProcessMiniBlocksFromMeV2 has been started")

	mapMiniBlocks := make(map[uint32]*block.MiniBlock)
	mapSCTxs := make(map[string]struct{})
	mapTxsForShard := make(map[uint32]int)
	mapScsForShard := make(map[uint32]int)

	numTxsAdded := 0
	numTxsBad := 0
	numTxsSkipped := 0
	numTxsFailed := 0
	numTxsWithInitialBalanceConsumed := 0
	numCrossShardScCallsOrSpecialTxs := 0

	totalTimeUsedForProcesss := time.Duration(0)
	totalTimeUsedForComputeGasConsumed := time.Duration(0)

	firstInvalidTxFound := false
	firstCrossShardScCallOrSpecialTxFound := false

	gasConsumedByMiniBlocksInSenderShard := uint64(0)
	mapGasConsumedByMiniBlockInReceiverShard := make(map[uint32]uint64)
	totalGasConsumedInSelfShard := txs.gasHandler.TotalGasConsumed()

	log.Debug("createAndProcessMiniBlocksFromMeV2", "totalGasConsumedInSelfShard", totalGasConsumedInSelfShard)

	senderAddressToSkip := []byte("")

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapMiniBlocks[shardID] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), shardID, block.TxBlock)
	}

	mapMiniBlocks[core.MetachainShardId] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), core.MetachainShardId, block.TxBlock)

	for index := range sortedTxs {
		if !haveTime() {
			log.Debug("time is out in createAndProcessMiniBlocksFromMeV2")
			break
		}

		tx, ok := sortedTxs[index].Tx.(*transaction.Transaction)
		if !ok {
			log.Debug("wrong type assertion",
				"hash", sortedTxs[index].TxHash,
				"sender shard", sortedTxs[index].SenderShardID,
				"receiver shard", sortedTxs[index].ReceiverShardID)
			continue
		}

		txHash := sortedTxs[index].TxHash
		senderShardID := sortedTxs[index].SenderShardID
		receiverShardID := sortedTxs[index].ReceiverShardID

		miniBlock, ok := mapMiniBlocks[receiverShardID]
		if !ok {
			log.Debug("miniblock is not created", "shard", receiverShardID)
			continue
		}

		isReceiverSmartContractAddress := core.IsSmartContractAddress(tx.RcvAddr)
		firstMiniBlockSplitForReceiverShardFound := (isReceiverSmartContractAddress && mapScsForShard[receiverShardID] == 0 && mapTxsForShard[receiverShardID] > 0) ||
			(!isReceiverSmartContractAddress && mapTxsForShard[receiverShardID] == 0 && mapScsForShard[receiverShardID] > 0)

		numNewMiniBlocks := 0
		if len(miniBlock.TxHashes) == 0 || firstCrossShardScCallOrSpecialTxFound {
			numNewMiniBlocks = 1
		}
		numNewTxs := 1

		isCrossShardScCallOrSpecialTx := receiverShardID != txs.shardCoordinator.SelfId() &&
			(isReceiverSmartContractAddress || len(tx.RcvUserName) > 0)
		if isCrossShardScCallOrSpecialTx {
			if !firstCrossShardScCallOrSpecialTxFound {
				numNewMiniBlocks++
			}
			numNewTxs += core.AdditionalScrForEachScCallOrSpecialTx
		}

		if isMaxBlockSizeReached(numNewMiniBlocks, numNewTxs) {
			log.Debug("max txs accepted in one block is reached",
				"num txs added", numTxsAdded,
				"total txs", len(sortedTxs))
			break
		}

		if isShardStuck != nil && isShardStuck(receiverShardID) {
			log.Trace("shard is stuck", "shard", receiverShardID)
			continue
		}

		if len(senderAddressToSkip) > 0 {
			if bytes.Equal(senderAddressToSkip, tx.GetSndAddr()) {
				numTxsSkipped++
				continue
			}
		}

		txMaxTotalCost := big.NewInt(0)
		isAddressSet := txs.balanceComputation.IsAddressSet(tx.GetSndAddr())
		if isAddressSet {
			txMaxTotalCost = txs.getTxMaxTotalCost(tx)
		}

		if isAddressSet {
			addressHasEnoughBalance := txs.balanceComputation.AddressHasEnoughBalance(tx.GetSndAddr(), txMaxTotalCost)
			if !addressHasEnoughBalance {
				numTxsWithInitialBalanceConsumed++
				continue
			}
		}

		snapshot := txs.accounts.JournalLen()

		gasConsumedByMiniBlockInReceiverShard := mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
		oldGasConsumedByMiniBlocksInSenderShard := gasConsumedByMiniBlocksInSenderShard
		oldGasConsumedByMiniBlockInReceiverShard := gasConsumedByMiniBlockInReceiverShard
		oldTotalGasConsumedInSelfShard := totalGasConsumedInSelfShard
		startTime := time.Now()
		err := txs.computeGasConsumed(
			senderShardID,
			receiverShardID,
			tx,
			txHash,
			&gasConsumedByMiniBlocksInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard,
			&totalGasConsumedInSelfShard)
		elapsedTime := time.Since(startTime)
		totalTimeUsedForComputeGasConsumed += elapsedTime
		if err != nil {
			log.Trace("createAndProcessMiniBlocksFromMeV2.computeGasConsumed", "error", err)
			continue
		}

		mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = gasConsumedByMiniBlockInReceiverShard

		// execute transaction to change the trie root hash
		startTime = time.Now()
		err = txs.processAndRemoveBadTransaction(
			txHash,
			tx,
			senderShardID,
			receiverShardID,
		)
		elapsedTime = time.Since(startTime)
		totalTimeUsedForProcesss += elapsedTime

		txs.mutAccountsInfo.Lock()
		txs.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
		txs.mutAccountsInfo.Unlock()

		if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
			if errors.Is(err, process.ErrHigherNonceInTransaction) {
				senderAddressToSkip = tx.GetSndAddr()
			}

			numTxsBad++
			log.Trace("bad tx",
				"error", err.Error(),
				"hash", txHash,
			)

			err = txs.accounts.RevertToSnapshot(snapshot)
			if err != nil {
				log.Warn("revert to snapshot", "error", err.Error())
			}

			txs.gasHandler.RemoveGasConsumed([][]byte{txHash})
			txs.gasHandler.RemoveGasRefunded([][]byte{txHash})

			gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
			mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = oldGasConsumedByMiniBlockInReceiverShard
			totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard
			continue
		}

		senderAddressToSkip = []byte("")

		gasRefunded := txs.gasHandler.GasRefunded(txHash)
		mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] -= gasRefunded
		if senderShardID == receiverShardID {
			gasConsumedByMiniBlocksInSenderShard -= gasRefunded
			totalGasConsumedInSelfShard -= gasRefunded
		}

		if errors.Is(err, process.ErrFailedTransaction) {
			if !firstInvalidTxFound {
				firstInvalidTxFound = true
				txs.blockSizeComputation.AddNumMiniBlocks(1)
			}

			txs.blockSizeComputation.AddNumTxs(1)
			numTxsFailed++
			continue
		}

		if isAddressSet {
			ok = txs.balanceComputation.SubBalanceFromAddress(tx.GetSndAddr(), txMaxTotalCost)
			if !ok {
				log.Error("createAndProcessMiniBlocksFromMeV2.SubBalanceFromAddress",
					"sender address", tx.GetSndAddr(),
					"tx max total cost", txMaxTotalCost,
					"err", process.ErrInsufficientFunds)
			}
		}

		if len(miniBlock.TxHashes) == 0 || firstMiniBlockSplitForReceiverShardFound {
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		if isReceiverSmartContractAddress {
			mapScsForShard[receiverShardID]++
			mapSCTxs[string(txHash)] = struct{}{}
		} else {
			mapTxsForShard[receiverShardID]++
		}

		miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
		txs.blockSizeComputation.AddNumTxs(1)
		if isCrossShardScCallOrSpecialTx {
			if !firstCrossShardScCallOrSpecialTxFound {
				firstCrossShardScCallOrSpecialTxFound = true
				txs.blockSizeComputation.AddNumMiniBlocks(1)
			}
			//we need to increment this as to account for the corresponding SCR hash
			txs.blockSizeComputation.AddNumTxs(core.AdditionalScrForEachScCallOrSpecialTx)
			numCrossShardScCallsOrSpecialTxs++
		}
		numTxsAdded++
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mapMiniBlocks, mapSCTxs)
	scheduledMiniBlocks := txs.createScheduledMiniBlocks(haveTime, isShardStuck, isMaxBlockSizeReached, sortedTxs, mapSCTxs)
	miniBlocks = append(miniBlocks, scheduledMiniBlocks...)

	log.Debug("createAndProcessMiniBlocksFromMeV2",
		"self shard", txs.shardCoordinator.SelfId(),
		"gas consumed in sender shard", gasConsumedByMiniBlocksInSenderShard,
		"total gas consumed in self shard", totalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard", mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID],
			"txs added", len(miniBlock.TxHashes),
			"not sc txs", mapTxsForShard[miniBlock.ReceiverShardID],
			"sc txs", mapScsForShard[miniBlock.ReceiverShardID])
	}

	log.Debug("createAndProcessMiniBlocksFromMeV2 has been finished",
		"total txs", len(sortedTxs),
		"num txs added", numTxsAdded,
		"num txs bad", numTxsBad,
		"num txs failed", numTxsFailed,
		"num txs skipped", numTxsSkipped,
		"num txs with initial balance consumed", numTxsWithInitialBalanceConsumed,
		"num cross shard sc calls or special txs", numCrossShardScCallsOrSpecialTxs,
		"used time for computeGasConsumed", totalTimeUsedForComputeGasConsumed,
		"used time for processAndRemoveBadTransaction", totalTimeUsedForProcesss)

	return miniBlocks, nil
}

func (txs *transactions) getMiniBlockSliceFromMapV2(
	mapMiniBlocks map[uint32]*block.MiniBlock,
	mapSCTxs map[string]struct{},
) block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		if miniBlock, ok := mapMiniBlocks[shardID]; ok {
			if len(miniBlock.TxHashes) > 0 {
				miniBlocks = append(miniBlocks, splitMiniBlockIfNeeded(miniBlock, mapSCTxs)...)
			}
		}
	}

	if miniBlock, ok := mapMiniBlocks[core.MetachainShardId]; ok {
		if len(miniBlock.TxHashes) > 0 {
			miniBlocks = append(miniBlocks, splitMiniBlockIfNeeded(miniBlock, mapSCTxs)...)
		}
	}

	return miniBlocks
}

func splitMiniBlockIfNeeded(miniBlock *block.MiniBlock, mapSCTxs map[string]struct{}) block.MiniBlockSlice {
	splitMiniBlocks := make(block.MiniBlockSlice, 0)
	notScTxHashes := make([][]byte, 0)
	scTxHashes := make([][]byte, 0)

	for _, txHash := range miniBlock.TxHashes {
		_, isSCTx := mapSCTxs[string(txHash)]
		if !isSCTx {
			notScTxHashes = append(notScTxHashes, txHash)
			continue
		}

		scTxHashes = append(scTxHashes, txHash)
	}

	if len(notScTxHashes) > 0 {
		notScMiniBlock := &block.MiniBlock{
			TxHashes:        notScTxHashes,
			SenderShardID:   miniBlock.SenderShardID,
			ReceiverShardID: miniBlock.ReceiverShardID,
			Type:            miniBlock.Type,
			Reserved:        miniBlock.Reserved,
		}

		splitMiniBlocks = append(splitMiniBlocks, notScMiniBlock)
	}

	if len(scTxHashes) > 0 {
		scMiniBlock := &block.MiniBlock{
			TxHashes:        scTxHashes,
			SenderShardID:   miniBlock.SenderShardID,
			ReceiverShardID: miniBlock.ReceiverShardID,
			Type:            miniBlock.Type,
			Reserved:        miniBlock.Reserved,
		}

		splitMiniBlocks = append(splitMiniBlocks, scMiniBlock)
	}

	return splitMiniBlocks
}

//TODO: Finish implementation of this method which should create all the scheduled mini blocks if needed
func (txs *transactions) createScheduledMiniBlocks(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
) block.MiniBlockSlice {
	log.Debug("createScheduledMiniBlocks has been started")

	mapMiniBlocks := make(map[uint32]*block.MiniBlock)

	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapMiniBlocks[shardID] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), shardID, block.ScheduledBlock)
	}

	mapMiniBlocks[core.MetachainShardId] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), core.MetachainShardId, block.ScheduledBlock)

	for index := range sortedTxs {
		if !haveTime() {
			log.Debug("time is out in createScheduledMiniBlocks")
			break
		}

		txHash := sortedTxs[index].TxHash
		senderShardID := sortedTxs[index].SenderShardID
		receiverShardID := sortedTxs[index].ReceiverShardID

		if isShardStuck != nil && isShardStuck(receiverShardID) {
			log.Trace("shard is stuck", "shard", receiverShardID)
			continue
		}

		tx, ok := sortedTxs[index].Tx.(*transaction.Transaction)
		if !ok {
			log.Debug("wrong type assertion",
				"hash", txHash,
				"sender shard", senderShardID,
				"receiver shard", receiverShardID)
			continue
		}

		if !core.IsSmartContractAddress(tx.RcvAddr) {
			continue
		}

		_, alreadyAdded := mapSCTxs[string(txHash)]
		if alreadyAdded {
			continue
		}

		log.Debug("this sc call could be added in scheduled mini blocks",
			"nonce", tx.Nonce,
			"hash", txHash,
			"sender shard", senderShardID,
			"receiver shard", receiverShardID)
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mapMiniBlocks, mapSCTxs)
	log.Debug("createScheduledMiniBlocks has been finished")
	return miniBlocks
}
