package txcache

import (
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
)

/*
Previously, when not-executable transactions were selected from the mempool, they were detected as bad in the processing flow.
There, they were collected and “targeted for deletion”. Then, deleted upon Node’s request.
Now, however, the not-executable ones remain in pool, until they are removed due to processing a transaction with equal or higher nonce.
Maybe we should implement a self-cleaning mechanism, in addition to the rarely-executing, generic tx pool cleaner.

Self-cleaning: mempool, please simulate a selection and remove all not-executable transactions that you hold.

System test - nonce-uri duplicate. spete cu scheduled, spete de concurență fină (receive in interceptor, accept, commit happens in processing flow, mempool insertion happens in interception flow)
*/

func (cache *TxCache) Cleanup(session SelectionSession, offset uint64, maxNum int, selectionLoopMaximumDuration time.Duration) (uint64) {
	logRemove.Debug(
		"TxCache.Cleanup: begin",
		"offset", offset,
		"maxNum", maxNum,
		"selectionLoopMaximumDuration", selectionLoopMaximumDuration,
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)
	return cache.RemoveSweepableTxs(session, offset, maxNum, selectionLoopMaximumDuration)
}

func (cache *TxCache) getOrderedSenders(offset uint64) []*txListForSender {
	senders := make([]*txListForSender, 0, cache.txListBySender.counter.Get())
	sender_addresses := cache.txListBySender.backingMap.Keys()
	sort.Slice(sender_addresses, func(i, j int) bool {
		return sender_addresses[i] < sender_addresses[j]
	})
	
	for _, sender := range shuffleSendersAddresses(sender_addresses,offset) {
		listForSender, ok := cache.txListBySender.backingMap.Get(sender)
		if ok {
			senders = append(senders, listForSender.(*txListForSender))
		}
	}
	return senders
}

func shuffleSendersAddresses(senders []string, offset uint64) []string {
	keys := make([]string, len(senders))
	mapSenders := make(map[string]string)
	var concat []byte

	randomness := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomness, offset)
	
	hasher := sha256.NewSha256()
	for i, v := range senders {
		concat = append([]byte(v), randomness...)

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

func (cache *TxCache) RemoveSweepableTxs(session SelectionSession, nonce uint64, maxNum int, selectionLoopMaximumDuration time.Duration) uint64 {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()
	selectionLoopStartTime := time.Now()
	evicted := make([][] byte, 0, cache.txByHash.counter.Get())

	senders := cache.getOrderedSenders(nonce)
	for _, sender := range senders{
		sessionWrapper := newSelectionSessionWrapper(session)
		sender_address := []byte (sender.sender)
		nonce:= sessionWrapper.getNonce(sender_address)-1

		if len(evicted) >= maxNum || time.Since(selectionLoopStartTime) > selectionLoopMaximumDuration{
			break
		}
		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(nonce)...)
	}
	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}
	logRemove.Debug("TxCache.Cleanup end", "block nonce", nonce, "len(cleaned)", len(evicted), "duration", time.Since(selectionLoopStartTime))
	return uint64(len(evicted))
}

func (listForSender *txListForSender) removeSweepableTransactionsReturnHashes(targetNonce uint64) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// We don't allow concurrent goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	lastTxNonceForDuplicatesProcessing:= targetNonce
	for element := listForSender.items.Front(); element != nil; {
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()
		txNonceForDuplicateProcessing:= lastTxNonceForDuplicatesProcessing
		lastTxNonceForDuplicatesProcessing = txNonce

		if txNonce > targetNonce && txNonce != txNonceForDuplicateProcessing{
			//fmt.Println("Skipped txNonce:", txNonce, ", targetNonce = ",targetNonce, " txNonceForDuplicateProcessing:", txNonceForDuplicateProcessing)	
			element = element.Next()
			continue
		}
		logRemove.Debug("TxCache.removeSweepableTransactionsReturnHashes",
			"txHash", tx.TxHash,
			"txNonce", txNonce,
			"targetNonce", targetNonce,
			"txNonceForDuplicateProcessing", txNonceForDuplicateProcessing,
		)
		fmt.Println("Not skipped txNonce:", txNonce, ", targetNonce = ",targetNonce, " txNonceForDuplicateProcessing:", txNonceForDuplicateProcessing)
		nextElement := element.Next()
		_ = listForSender.items.Remove(element)
		listForSender.onRemovedListElement(element)
		element = nextElement

		// Keep track of removed transactions
		evictedTxHashes = append(evictedTxHashes, tx.TxHash)
	}

	return evictedTxHashes
}