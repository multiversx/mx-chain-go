package testscommon

// StateStatisticsHandlerStub -
type StateStatisticsHandlerStub struct {
	ResetCalled                 func()
	ResetSnapshotCalled         func()
	IncrCacheCalled             func()
	CacheCalled                 func() uint64
	IncrSnapshotCacheCalled     func()
	SnapshotCacheCalled         func() uint64
	IncrPersisterCalled         func(epoch uint32)
	PersisterCalled             func(epoch uint32) uint64
	IncrSnapshotPersisterCalled func(epoch uint32)
	SnapshotPersisterCalled     func(epoch uint32) uint64
	IncrTrieCalled              func()
	TrieCalled                  func() uint64
	ProcessingStatsCalled       func() []string
	SnapshotStatsCalled         func() []string
}

// Reset -
func (stub *StateStatisticsHandlerStub) Reset() {
	if stub.ResetCalled != nil {
		stub.ResetCalled()
	}
}

// ResetSnapshot -
func (stub *StateStatisticsHandlerStub) ResetSnapshot() {
	if stub.ResetSnapshotCalled != nil {
		stub.ResetSnapshotCalled()
	}
}

// IncrCache -
// TODO: replace Incr with Increment on all usages in this file + rename the interface and the other 2 implementations
func (stub *StateStatisticsHandlerStub) IncrCache() {
	if stub.IncrCacheCalled != nil {
		stub.IncrCacheCalled()
	}
}

// Cache -
func (stub *StateStatisticsHandlerStub) Cache() uint64 {
	if stub.CacheCalled != nil {
		return stub.CacheCalled()
	}

	return 0
}

// IncrSnapshotCache -
func (stub *StateStatisticsHandlerStub) IncrSnapshotCache() {
	if stub.IncrSnapshotCacheCalled != nil {
		stub.IncrSnapshotCacheCalled()
	}
}

// SnapshotCache -
func (stub *StateStatisticsHandlerStub) SnapshotCache() uint64 {
	if stub.SnapshotCacheCalled != nil {
		return stub.SnapshotCacheCalled()
	}

	return 0
}

// IncrPersister -
func (stub *StateStatisticsHandlerStub) IncrPersister(epoch uint32) {
	if stub.IncrPersisterCalled != nil {
		stub.IncrPersisterCalled(epoch)
	}
}

// Persister -
func (stub *StateStatisticsHandlerStub) Persister(epoch uint32) uint64 {
	if stub.PersisterCalled != nil {
		return stub.PersisterCalled(epoch)
	}

	return 0
}

// IncrSnapshotPersister -
func (stub *StateStatisticsHandlerStub) IncrSnapshotPersister(epoch uint32) {
	if stub.IncrSnapshotPersisterCalled != nil {
		stub.IncrSnapshotPersisterCalled(epoch)
	}
}

// SnapshotPersister -
func (stub *StateStatisticsHandlerStub) SnapshotPersister(epoch uint32) uint64 {
	if stub.SnapshotPersisterCalled != nil {
		return stub.SnapshotPersisterCalled(epoch)
	}

	return 0
}

// IncrTrie -
func (stub *StateStatisticsHandlerStub) IncrTrie() {
	if stub.IncrTrieCalled != nil {
		stub.IncrTrieCalled()
	}
}

// Trie -
func (stub *StateStatisticsHandlerStub) Trie() uint64 {
	if stub.TrieCalled != nil {
		return stub.TrieCalled()
	}

	return 0
}

// ProcessingStats -
func (stub *StateStatisticsHandlerStub) ProcessingStats() []string {
	if stub.ProcessingStatsCalled != nil {
		return stub.ProcessingStatsCalled()
	}

	return make([]string, 0)
}

// SnapshotStats -
func (stub *StateStatisticsHandlerStub) SnapshotStats() []string {
	if stub.SnapshotStatsCalled != nil {
		return stub.SnapshotStatsCalled()
	}

	return make([]string, 0)
}

// IsInterfaceNil -
func (stub *StateStatisticsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
