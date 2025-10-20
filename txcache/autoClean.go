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

	cleanupLoopStartTime := time.Now()

	logRemove.Debug("TxCache.RemoveSweepableTxs start",
		"randomness", randomness,
		"maxNum", maxNum,
		"cleanupLoopMaximumDuration", cleanupLoopMaximumDurationMs,
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

		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(targetNonce)...)
	}

	if len(evicted) > 0 {
		txs := cache.txByHash.GetTxsBulk(evicted)

		untrackedEvicted := cache.tracker.GetBulkOfUntrackedTransactions(txs)
		cache.txByHash.RemoveTxsBulk(untrackedEvicted)
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

func (listForSender *txListForSender) removeSweepableTransactionsReturnHashes(targetNonce uint64) [][]byte {
	txHashesToEvict := make([][]byte, 0)

	// We don't allow concurrent goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	for element := listForSender.items.Front(); element != nil; {
		// finds transactions with lower nonces
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()

		// nonces are sorted ascending, so we can stop as soon as we find a nonce that is higher
		if txNonce > targetNonce {
			break
		}

		logRemove.Debug("TxCache.removeSweepableTransactionsReturnHashes",
			"txHash", tx.TxHash,
			"txNonce", txNonce,
			"targetNonce", targetNonce,
		)

		nextElement := element.Next()
		_ = listForSender.items.Remove(element)
		listForSender.onRemovedListElement(element)
		element = nextElement

		// Keep track of removed transactions
		txHashesToEvict = append(txHashesToEvict, tx.TxHash)
	}

	return txHashesToEvict
}
