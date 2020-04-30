package txcache

import (
	"sort"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/storage/txcache/maps"
)

const numberOfScoreChunks = uint32(100)

// txListBySenderMap is a map-like structure for holding and accessing transactions by sender
type txListBySenderMap struct {
	backingMap  *maps.BucketSortedMap
	cacheConfig CacheConfig
	counter     atomic.Counter
}

// newTxListBySenderMap creates a new instance of TxListBySenderMap
func newTxListBySenderMap(nChunksHint uint32, cacheConfig CacheConfig) txListBySenderMap {
	backingMap := maps.NewBucketSortedMap(nChunksHint, numberOfScoreChunks)

	return txListBySenderMap{
		backingMap:  backingMap,
		cacheConfig: cacheConfig,
	}
}

// addTx adds a transaction in the map, in the corresponding list (selected by its sender)
func (txMap *txListBySenderMap) addTx(tx *WrappedTransaction) {
	sender := string(tx.Tx.GetSndAddr())
	listForSender := txMap.getOrAddListForSender(sender)
	listForSender.AddTx(tx)
	txMap.notifyScoreChange(listForSender)
}

func (txMap *txListBySenderMap) notifyScoreChange(txList *txListForSender) {
	txMap.backingMap.NotifyScoreChange(txList)
}

func (txMap *txListBySenderMap) getOrAddListForSender(sender string) *txListForSender {
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		listForSender = txMap.addSender(sender)
	}

	return listForSender
}

func (txMap *txListBySenderMap) getListForSender(sender string) (*txListForSender, bool) {
	listForSenderUntyped, ok := txMap.backingMap.Get(sender)
	if !ok {
		return nil, false
	}

	listForSender := listForSenderUntyped.(*txListForSender)
	return listForSender, true
}

func (txMap *txListBySenderMap) addSender(sender string) *txListForSender {
	listForSender := newTxListForSender(sender, &txMap.cacheConfig)

	txMap.backingMap.Set(listForSender)
	txMap.counter.Increment()

	return listForSender
}

// removeTx removes a transaction from the map
func (txMap *txListBySenderMap) removeTx(tx *WrappedTransaction) bool {
	sender := string(tx.Tx.GetSndAddr())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		txMap.onRemoveTxInconsistency(sender)
		return false
	}

	isFound := listForSender.RemoveTx(tx)

	if listForSender.IsEmpty() {
		txMap.removeSender(sender)
	} else {
		txMap.notifyScoreChange(listForSender)
	}

	return isFound
}

func (txMap *txListBySenderMap) removeSender(sender string) {
	if !txMap.backingMap.Has(sender) {
		return
	}

	txMap.backingMap.Remove(sender)
	txMap.counter.Decrement()
}

// RemoveSendersBulk removes senders, in bulk
func (txMap *txListBySenderMap) RemoveSendersBulk(senders []string) uint32 {
	oldCount := uint32(txMap.counter.Get())

	for _, senderKey := range senders {
		txMap.removeSender(senderKey)
	}

	newCount := uint32(txMap.counter.Get())
	nRemoved := oldCount - newCount
	return nRemoved
}

func (txMap *txListBySenderMap) notifyAccountNonce(accountKey []byte, nonce uint64) {
	sender := string(accountKey)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return
	}

	listForSender.notifyAccountNonce(nonce)
}

func (txMap *txListBySenderMap) getSnapshotAscendingWithDeterministicallySortedHead() []*txListForSender {
	snapshot := txMap.getSnapshotAscending()
	if len(snapshot) == 0 {
		return snapshot
	}

	sorter := func(i, j int) bool {
		senderI := snapshot[i]
		senderJ := snapshot[j]

		delta := int(senderI.getLastComputedScore()) - int(senderJ.getLastComputedScore())
		if delta == 0 {
			delta = strings.Compare(senderI.GetKey(), senderJ.GetKey())
		}

		return delta < 0
	}

	headSize := getSnapshotHeadSize(snapshot)
	sort.Slice(snapshot[:headSize], sorter)
	return snapshot
}

func getSnapshotHeadSize(snapshot []*txListForSender) int {
	headSize := core.MinInt(len(snapshot), maxSendersSnapshotHeadSize)

	i := 0
	previousScore := uint32(0)
	currentScoreStart := 0
	currentScore := uint32(0)

	for {
		currentScore = snapshot[i].getLastComputedScore()
		if currentScore != previousScore {
			currentScoreStart = i
		}

		i++

		if i >= headSize {
			headSize = currentScoreStart
			break
		}
	}

	return headSize
}

func (txMap *txListBySenderMap) getSnapshotAscending() []*txListForSender {
	itemsSnapshot := txMap.backingMap.GetSnapshotAscending()
	listsSnapshot := make([]*txListForSender, len(itemsSnapshot))

	for i, item := range itemsSnapshot {
		listsSnapshot[i] = item.(*txListForSender)
	}

	return listsSnapshot
}

func (txMap *txListBySenderMap) getSnapshotDescending() []*txListForSender {
	itemsSnapshot := txMap.backingMap.GetSnapshotDescending()
	listsSnapshot := make([]*txListForSender, len(itemsSnapshot))

	for i, item := range itemsSnapshot {
		listsSnapshot[i] = item.(*txListForSender)
	}

	return listsSnapshot
}

func (txMap *txListBySenderMap) getSnapshotDescendingWithDeterministicallySortedTail() []*txListForSender {
	snapshot := txMap.getSnapshotAscendingWithDeterministicallySortedHead()

	for i, j := 0, len(snapshot)-1; i < j; i, j = i+1, j-1 {
		snapshot[i], snapshot[j] = snapshot[j], snapshot[i]
	}

	return snapshot
}

func (txMap *txListBySenderMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}
