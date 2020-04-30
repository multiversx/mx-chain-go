package txcache

import (
	"sort"
	"strings"
)

// getSnapshotAscendingWithDeterministicallySortedHead returns a snapshot of senders,
// where the senders with the lowest scores (the head of the snapshot) are further sorted lexicographically (by their address).
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

	// We only sort the head of the snapshot
	headSize := getSnapshotHeadSize(snapshot)
	sort.Slice(snapshot[:headSize], sorter)
	return snapshot
}

// getSnapshotHeadSize computes the size of a snapshot's head, that is, the number of elements
// in the snapshot to further sort.
// This size defaults to "defaultSendersSnapshotHeadSize" and then is corrected to span over "complete" score buckets ("ceiling function").
func getSnapshotHeadSize(snapshot []*txListForSender) int {
	if len(snapshot) < defaultSendersSnapshotHeadSize {
		return len(snapshot)
	}

	// The default
	headSize := defaultSendersSnapshotHeadSize

	// The correction
	ceilingScore := snapshot[headSize-1].getLastComputedScore()
	for i := headSize; i < len(snapshot); i++ {
		if snapshot[i].getLastComputedScore() != ceilingScore {
			break
		}

		headSize++
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
