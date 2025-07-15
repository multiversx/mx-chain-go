package txcache

import (
	"bytes"
	"encoding/binary"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
)

// Cleanup simulates a selection and removes not-executable transactions. Initial implementation: lower and duplicate nonces
func (cache *TxCache) Cleanup(session SelectionSession, randomness uint64, maxNum int, removalLoopMaximumDurationMs time.Duration) (uint64) {
	logRemove.Debug(
		"TxCache.Cleanup: begin",
		"randomness", randomness,
		"maxNum", maxNum,
		"removalLoopMaximumDuration", removalLoopMaximumDurationMs,
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)
	return cache.RemoveSweepableTxs(session, randomness, maxNum, removalLoopMaximumDurationMs)
}

func (cache *TxCache) getDeterministicallyShuffledSenders(randomness uint64) []*txListForSender {
	senders := make([]*txListForSender, 0, cache.txListBySender.counter.Get())
	senderAddresses := cache.txListBySender.backingMap.Keys()
	
	shuffledSenderAdresses:= shuffleSendersAddresses(senderAddresses, randomness)
	for _, sender := range shuffledSenderAdresses {
		listForSender, ok := cache.txListBySender.backingMap.Get(sender)
		if ok {
			senders = append(senders, listForSender.(*txListForSender))
		}
	}
	
	return senders
}

type addressShufflingItem struct {  
    address      string  
    shufflingKey []byte  
}  

func shuffleSendersAddresses(senders []string, randomness uint64) []string {  
    randomnessAsBytes := make([]byte, 8)  
    binary.LittleEndian.PutUint64(randomnessAsBytes, randomness)  

    hasher := sha256.NewSha256()  
    items := make([]*addressShufflingItem, 0, len(senders))  
    shuffledAddresses := make([]string, len(senders))  

    for _, address := range senders {  
        addressWithRandomness := append([]byte(address), randomnessAsBytes...)  
        shufflingKey := hasher.Compute(string(addressWithRandomness))  

        items = append(items, &addressShufflingItem{  
            address:      address,  
            shufflingKey: shufflingKey,  
        })  
    }  

    // Deterministic shuffling.  
	sort.Slice(items, func(i, j int) bool {
		cmp := bytes.Compare(items[i].shufflingKey, items[j].shufflingKey)
		if cmp != 0 {
			return cmp > 0
		}
		return items[i].address > items[j].address
	})

    for i := 0; i < len(senders); i++ {  
        shuffledAddresses[i] = items[i].address  
    }  

    return shuffledAddresses  
} 

func (cache *TxCache) RemoveSweepableTxs(session SelectionSession, randomness uint64, maxNum int, removalLoopMaximumDuration time.Duration) uint64 {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()
	
	removalLoopStartTime := time.Now()
	evicted := make([][] byte, 0, cache.txByHash.counter.Get())

	sessionWrapper := newSelectionSessionWrapper(session)
	senders := cache.getDeterministicallyShuffledSenders(randomness)
	
	for _, sender := range senders{
		senderAddress := []byte (sender.sender)
		lastCommittedNonce:= sessionWrapper.getNonce(senderAddress) - 1

		if len(evicted) >= maxNum || time.Since(removalLoopStartTime) > removalLoopMaximumDuration{
			break
		}
		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(lastCommittedNonce)...)
	}
	
	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}
	
	logRemove.Debug("TxCache.RemoveSweepableTxs end", "randomness", randomness, "len(evicted)", len(evicted), "duration", time.Since(removalLoopStartTime))
	
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
			"targetNonce", targetNonce,
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
