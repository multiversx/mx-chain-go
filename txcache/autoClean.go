package txcache

import (
	"bytes"
	"encoding/binary"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
)

// Cleanup simulates a selection and removes not-executable transactions. Initial implementation: lower nonces
// TODO Maybe we can think of an alternative fast and simple sort & shuffle at the same time. Maybe we can do a single sorting (in a separate PR).
func (cache *TxCache) Cleanup(session SelectionSession, randomness uint64, maxNum int, cleanupLoopMaximumDurationMs time.Duration) uint64 {
	logRemove.Debug(
		"TxCache.Cleanup: begin",
		"randomness", randomness,
		"maxNum", maxNum,
		"cleanupLoopMaximumDuration", cleanupLoopMaximumDurationMs,
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)
	return cache.RemoveSweepableTxs(session, randomness, maxNum, cleanupLoopMaximumDurationMs)
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

func (cache *TxCache) RemoveSweepableTxs(session SelectionSession, randomness uint64, maxNum int, cleanupLoopMaximumDurationMs time.Duration) uint64 {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()

	cleanupLoopStartTime := time.Now()
	evicted := make([][]byte, 0, cache.txByHash.counter.Get())

	virtualSession := newVirtualSelectionSession(session)
	senders := cache.getDeterministicallyShuffledSenders(randomness)

	for _, sender := range senders {
		senderAddress := []byte(sender.sender)

		lastCommittedNonce, err := virtualSession.getNonce(senderAddress)
		if err != nil {
			log.Debug("TxCache.RemoveSweepableTxs",
				"address", senderAddress,
				"err", err,
			)
			continue
		}

		// we want to remove transactions with nonces < lastCommittedNonce
		if lastCommittedNonce == 0 {
			continue
		}

		lastCommittedNonce -= 1

		if len(evicted) >= maxNum || time.Since(cleanupLoopStartTime) > cleanupLoopMaximumDurationMs {
			break
		}
		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(lastCommittedNonce)...)
	}

	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}

	logRemove.Debug("TxCache.RemoveSweepableTxs end", "randomness", randomness, "len(evicted)", len(evicted), "duration", time.Since(cleanupLoopStartTime))

	return uint64(len(evicted))
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

		if txNonce > targetNonce {
			element = element.Next()
			continue
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
