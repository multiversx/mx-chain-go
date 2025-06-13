package txcache

import (
	"encoding/binary"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
)

// Cleanup simulates a selection and removes not-executable transactions. Initial implementation: lower and duplicate nonces
func (cache *TxCache) Cleanup(session SelectionSession, randomness uint64, maxNum int, removalLoopMaximumDuration time.Duration) (uint64) {
	logRemove.Debug(
		"TxCache.Cleanup: begin",
		"randomness", randomness,
		"maxNum", maxNum,
		"removalLoopMaximumDuration", removalLoopMaximumDuration,
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)
	return cache.RemoveSweepableTxs(session, randomness, maxNum, removalLoopMaximumDuration)
}

func (cache *TxCache) getOrderedSenders(randomness uint64) []*txListForSender {
	senders := make([]*txListForSender, 0, cache.txListBySender.counter.Get())
	sender_addresses := cache.txListBySender.backingMap.Keys()
	sort.Slice(sender_addresses, func(i, j int) bool {
		return sender_addresses[i] < sender_addresses[j]
	})
	
	for _, sender := range shuffleSendersAddresses(sender_addresses,randomness) {
		listForSender, ok := cache.txListBySender.backingMap.Get(sender)
		if ok {
			senders = append(senders, listForSender.(*txListForSender))
		}
	}
	
	return senders
}

func shuffleSendersAddresses(senders []string, randomness uint64) []string {
	keys := make([]string, len(senders))
	mapSenders := make(map[string]string)
	var concat []byte

	adressRandomness := make([]byte, 8)
	binary.LittleEndian.PutUint64(adressRandomness, randomness)
	
	hasher := sha256.NewSha256()
	for i, v := range senders {
		concat = append([]byte(v), adressRandomness...)

		keys[i] = string(hasher.Compute(string(concat)))
		mapSenders[keys[i]] = v
	}

	sort.Strings(keys)

	result := make([]string, len(senders))
	for i := 0; i < len(senders); i++ {
		result[i] = mapSenders[keys[i]]
	}

	return result
}

func (cache *TxCache) RemoveSweepableTxs(session SelectionSession, randomness uint64, maxNum int, removalLoopMaximumDuration time.Duration) uint64 {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()
	
	removalLoopStartTime := time.Now()
	evicted := make([][] byte, 0, cache.txByHash.counter.Get())

	sessionWrapper := newSelectionSessionWrapper(session)
	senders := cache.getOrderedSenders(randomness)
	
	for _, sender := range senders{
		senderAddress := []byte (sender.sender)
		lastCommitedNonce:= sessionWrapper.getNonce(senderAddress) - 1

		if len(evicted) >= maxNum || time.Since(removalLoopStartTime) > removalLoopMaximumDuration{
			break
		}
		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(lastCommitedNonce)...)
	}
	
	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}
	
	logRemove.Debug("TxCache.Cleanup end", "randomness", randomness, "len(cleaned)", len(evicted), "duration", time.Since(removalLoopStartTime))
	
	return uint64(len(evicted))
}

func (listForSender *txListForSender) removeSweepableTransactionsReturnHashes(targetNonce uint64) [][]byte {
	txHashesToEvict := make([][]byte, 0)

	// We don't allow concurrent goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	lastTxNonceForDuplicatesProcessing:= targetNonce
	
	for element := listForSender.items.Front(); element != nil; {
		// finds transactions with lower nonces, finds transactions with duplicate nonces
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()

		// set the current nonce to be checked for duplicate and remember last transaction nonce that has been encountered
		txNonceForDuplicateProcessing:= lastTxNonceForDuplicatesProcessing
		lastTxNonceForDuplicatesProcessing = txNonce

		if txNonce > targetNonce && txNonce != txNonceForDuplicateProcessing{
			element = element.Next()
			continue
		}
		
		logRemove.Debug("TxCache.removeSweepableTransactionsReturnHashes",
			"txHash", tx.TxHash,
			"txNonce", txNonce,
			"lastCommitedTxNonce", targetNonce,
			"txNonceForDuplicateProcessing", txNonceForDuplicateProcessing,
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
