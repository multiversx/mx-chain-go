package goroutines

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common/goroutines"
)

type goCounter struct {
	mutSnapshots      sync.RWMutex
	snapshots         []*snapshot
	goRoutinesFetcher func() string
	filter            func(goRoutineData string) bool
}

// NewGoCounter returns an instance of the go counter
func NewGoCounter(filterFunc func(goRoutineData string) bool) *goCounter {
	if filterFunc == nil {
		panic("NewGoCounter can not work with a nil filter")
	}

	return &goCounter{
		snapshots: make([]*snapshot, 0),
		filter:    filterFunc,
		goRoutinesFetcher: func() string {
			return goroutines.GetGoRoutines()
		},
	}
}

// Reset will clean all stored snapshots
func (gc *goCounter) Reset() {
	gc.mutSnapshots.Lock()
	gc.snapshots = make([]*snapshot, 0)
	gc.mutSnapshots.Unlock()
}

// DiffGoRoutines makes an absolute difference* between first and second snapshots indexes.
// If one of the index is invalid (-1, > maximum snapshots) the difference will be done against a nil snapshot
// this difference will return any stored routines info from the non-nil instance.
// If both the indexes do not have a corresponding snapshot, this function will return an empty slice
// * absolute difference = (snapshot1 - snapshot2) U (snapshot2 - snapshot1)
func (gc *goCounter) DiffGoRoutines(firstSnapshotIdx, secondSnapshotIdx int) []*RoutineInfo {
	gc.mutSnapshots.RLock()
	defer gc.mutSnapshots.RUnlock()

	snapshot1 := gc.getSnapshot(firstSnapshotIdx)
	snapshot2 := gc.getSnapshot(secondSnapshotIdx)
	if snapshot1 == nil && snapshot2 == nil {
		return make([]*RoutineInfo, 0)
	}
	if snapshot1 != nil {
		return snapshot1.diff(snapshot2)
	}

	return snapshot2.diff(snapshot1)
}

func (gc *goCounter) getSnapshot(idx int) *snapshot {
	if idx < 0 || idx >= len(gc.snapshots) {
		return nil
	}

	return gc.snapshots[idx]
}

// Snapshot will make a snapshot of the go routines that will pass through the filter function
func (gc *goCounter) Snapshot() (int, error) {
	goRoutinesData := gc.goRoutinesFetcher()
	s, err := newSnapshot(goRoutinesData, gc.filter)
	if err != nil {
		return 0, err
	}

	gc.mutSnapshots.Lock()
	defer gc.mutSnapshots.Unlock()

	gc.snapshots = append(gc.snapshots, s)

	return len(gc.snapshots) - 1, nil
}
