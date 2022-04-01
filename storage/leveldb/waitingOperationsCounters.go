package leveldb

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ElrondNetwork/elrond-go/common"
)

// waitingOperationsCounters is not concurrent safe
type waitingOperationsCounters struct {
	counters           map[common.StorageAccessType]int
	numNonZeroCounters int
}

func newWaitingOperationsCounters() *waitingOperationsCounters {
	return &waitingOperationsCounters{
		counters: make(map[common.StorageAccessType]int),
	}
}

func (woc *waitingOperationsCounters) increment(priority common.StorageAccessType) {
	old := woc.counters[priority]
	woc.counters[priority]++
	if old == 0 {
		woc.numNonZeroCounters++
	}
}

func (woc *waitingOperationsCounters) decrement(priority common.StorageAccessType) {
	old := woc.counters[priority]
	if old < 1 {
		return
	}
	if old == 1 {
		woc.numNonZeroCounters--
	}

	woc.counters[priority]--
}

func (woc *waitingOperationsCounters) snapshotString() string {
	keys := make([]common.StorageAccessType, 0, len(woc.counters))
	for priority := range woc.counters {
		keys = append(keys, priority)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	countersStrings := make([]string, 0, len(woc.counters))
	for _, prio := range keys {
		countersStrings = append(countersStrings, fmt.Sprintf("%s: %d", prio, woc.counters[prio]))
	}

	return strings.Join(countersStrings, ", ")
}
