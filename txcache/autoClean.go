package txcache

import (
	"bytes"
	"encoding/binary"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-go/common"
)

// Cleanup simulates a selection and removes not-executable transactions. Initial implementation: lower nonces
func (cache *TxCache) Cleanup(accountsProvider common.AccountNonceProvider, randomness uint64, maxNum int, cleanupLoopMaximumDurationMs time.Duration) uint64 {
	logRemove.Debug(
		"TxCache.Cleanup: begin",
		"randomness", randomness,
		"maxNum", maxNum,
		"cleanupLoopMaximumDuration", cleanupLoopMaximumDurationMs,
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)
	return cache.RemoveSweepableTxs(accountsProvider, randomness, maxNum, cleanupLoopMaximumDurationMs)
}

func (cache *TxCache) RemoveSweepableTxs(accountsProvider common.AccountNonceProvider, randomness uint64, maxNum int, cleanupLoopMaximumDurationMs time.Duration) uint64 {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()

	rootHash, err := accountsProvider.GetRootHash()
	if err != nil {
		log.Debug("TxCache.RemoveSweepableTxs: failed to get root hash", "err", err)
	}

	cleanupLoopStartTime := time.Now()

	logRemove.Debug("TxCache.RemoveSweepableTxs start",
		"randomness", randomness,
		"maxNum", maxNum,
		"cleanupLoopMaximumDuration", cleanupLoopMaximumDurationMs,
		"rootHash", rootHash,
	)

	evicted := make([][]byte, 0, cache.txByHash.counter.Get())

	senders := cache.getDeterministicallyShuffledSenders(randomness)

	for _, sender := range senders {
		senderAddress := []byte(sender.sender)

		accountNonce, _, err := accountsProvider.GetAccountNonce(senderAddress)
		if err != nil {
			log.Debug("TxCache.RemoveSweepableTxs",
				"address", senderAddress,
				"err", err,
			)
			continue
		}

		// nothing to do
		if accountNonce == 0 {
			continue
		}

		// stop if we reached the max number of evicted transactions for this cleanup loop
		if len(evicted) >= maxNum {
			logRemove.Debug("TxCache.RemoveSweepableTxs reached maxNum",
				"len(evicted)", len(evicted),
				"maxNum", maxNum,
			)
			break
		}

		// stop if we reached the maximum duration for this cleanup loop
		if time.Since(cleanupLoopStartTime) > cleanupLoopMaximumDurationMs {
			logRemove.Debug("TxCache.RemoveSweepableTxs reached cleanupLoopMaximumDuration",
				"len(evicted)", len(evicted),
				"duration", time.Since(cleanupLoopStartTime),
				"cleanupLoopMaximumDuration", cleanupLoopMaximumDurationMs,
			)
			break
		}

		// we want to remove transactions with nonces < lastCommittedNonce
		targetNonce := accountNonce - 1

		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(targetNonce, cache.tracker)...)
	}

	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}

	logRemove.Debug("TxCache.RemoveSweepableTxs end",
		"randomness", randomness,
		"len(evicted)", len(evicted),
		"duration", time.Since(cleanupLoopStartTime),
	)

	return uint64(len(evicted))
}

func (cache *TxCache) getDeterministicallyShuffledSenders(randomness uint64) []*txListForSender {
	senders := make([]*txListForSender, 0, cache.txListBySender.counter.Get())
	senderAddresses := cache.txListBySender.backingMap.Keys()

	shuffleSendersAddresses(senderAddresses, randomness)
	for _, sender := range senderAddresses {
		listForSender, ok := cache.txListBySender.backingMap.Get(sender)
		if ok {
			senders = append(senders, listForSender.(*txListForSender))
		}
	}

	return senders
}

func shuffleSendersAddresses(senders []string, randomness uint64) {
	randomnessAsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomnessAsBytes, randomness)

	hasher := sha256.NewSha256()
	keys := make([][]byte, len(senders))

	for i, addr := range senders {
		addrWithRand := append([]byte(addr), randomnessAsBytes...)
		keys[i] = hasher.Compute(string(addrWithRand))
	}

	sort.Slice(senders, func(i, j int) bool {
		cmp := bytes.Compare(keys[i], keys[j])
		if cmp != 0 {
			return cmp > 0
		}
		return senders[i] > senders[j]
	})
}

func (listForSender *txListForSender) removeSweepableTransactionsReturnHashes(targetNonce uint64, tracker *selectionTracker) [][]byte {
	txHashesToEvict := make([][]byte, 0)

	// We don't allow concurrent goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	// We iterate from start to end (ascending nonce)
	// We want to remove transactions with nonce < targetNonce ONLY if they are NOT tracked?
	// The original code says: "finds transactions with lower nonces ... txNonce > targetNonce { break } ... if !tracker.IsTracked { Remove }"
	// So we are removing untracked transactions that are "old" (nonce <= targetNonce).

	// Since we are modifying the slice in-place, we should be careful.
	// Filter approach: keep transactions that SHOULD stay.

	// OR reuse existing slice and compact? Reuse is better for allocs.
	// But simple logic first.

	// We only need to check up to where nonce > targetNonce.

	cutoffIndex := len(listForSender.items)
	for i, tx := range listForSender.items {
		if tx.Tx.GetNonce() > targetNonce {
			cutoffIndex = i
			break
		}
	}

	// Items from [0...cutoffIndex-1] are candidates for removal.
	// Items from [cutoffIndex...] remain untouched.

	// We will rebuild the prefix [0...cutoffIndex]
	keptItems := make([]*WrappedTransaction, 0, cutoffIndex)

	for i := 0; i < cutoffIndex; i++ {
		tx := listForSender.items[i]

		shouldRemove := !tracker.IsTransactionTracked(tx)
		// Logic check: original code removed if !IsTransactionTracked.

		if shouldRemove {
			logRemove.Debug("TxCache.removeSweepableTransactionsReturnHashes",
				"txHash", tx.TxHash,
				"txNonce", tx.Tx.GetNonce(),
				"targetNonce", targetNonce,
			)
			txHashesToEvict = append(txHashesToEvict, tx.TxHash)
			listForSender.onRemovedTransaction(tx)
		} else {
			keptItems = append(keptItems, tx)
		}
	}

	// Reassemble: keptItems + remaining items
	// We can reuse listForSender.items array if careful, but append is safe.
	// Optimization:
	// copy remaining to keptItems
	keptItems = append(keptItems, listForSender.items[cutoffIndex:]...)

	// Replace
	// Nil out old pointers for GC safety if we are shrinking?
	// The new items slice will overwrite.
	listForSender.items = keptItems

	return txHashesToEvict
}
