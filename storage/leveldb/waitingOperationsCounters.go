package leveldb

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/common"
)

type waitingOperationsCounters struct {
	mut      sync.RWMutex
	counters map[common.StorageAccessType]int
}

func newWaitingOperationsCounters() *waitingOperationsCounters {
	return &waitingOperationsCounters{
		counters: make(map[common.StorageAccessType]int),
	}
}

func (woc *waitingOperationsCounters) getCounter(priority common.StorageAccessType) int {
	woc.mut.RLock()
	defer woc.mut.RUnlock()

	return woc.counters[priority]
}

func (woc *waitingOperationsCounters) increment(priority common.StorageAccessType) {
	woc.mut.Lock()
	defer woc.mut.Unlock()

	woc.counters[priority]++
}

func (woc *waitingOperationsCounters) decrement(priority common.StorageAccessType) {
	woc.mut.Lock()
	defer woc.mut.Unlock()

	woc.counters[priority]--
}

func (woc *waitingOperationsCounters) snapshotString() string {
	woc.mut.RLock()
	defer woc.mut.RUnlock()

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
