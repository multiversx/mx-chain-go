package testscommon

// StateStatisticsHandlerStub -
type StateStatisticsHandlerStub struct {
	ResetCalled                      func()
	ResetSnapshotCalled              func()
	IncrementCacheCalled             func()
	CacheCalled                      func() uint64
	IncrementSnapshotCacheCalled     func()
	SnapshotCacheCalled              func() uint64
	IncrementPersisterCalled         func(epoch uint32)
	PersisterCalled                  func(epoch uint32) uint64
	IncrementSnapshotPersisterCalled func(epoch uint32)
	SnapshotPersisterCalled          func(epoch uint32) uint64
	IncrementTrieCalled              func()
	TrieCalled                       func() uint64
	ProcessingStatsCalled            func() []string
	SnapshotStatsCalled              func() []string
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

// IncrementCache -
func (stub *StateStatisticsHandlerStub) IncrementCache() {
	if stub.IncrementCacheCalled != nil {
		stub.IncrementCacheCalled()
	}
}

// Cache -
func (stub *StateStatisticsHandlerStub) Cache() uint64 {
	if stub.CacheCalled != nil {
		return stub.CacheCalled()
	}

	return 0
}

// IncrementSnapshotCache -
func (stub *StateStatisticsHandlerStub) IncrementSnapshotCache() {
	if stub.IncrementSnapshotCacheCalled != nil {
		stub.IncrementSnapshotCacheCalled()
	}
}

// SnapshotCache -
func (stub *StateStatisticsHandlerStub) SnapshotCache() uint64 {
	if stub.SnapshotCacheCalled != nil {
		return stub.SnapshotCacheCalled()
	}

	return 0
}

// IncrementPersister -
func (stub *StateStatisticsHandlerStub) IncrementPersister(epoch uint32) {
	if stub.IncrementPersisterCalled != nil {
		stub.IncrementPersisterCalled(epoch)
	}
}

// Persister -
func (stub *StateStatisticsHandlerStub) Persister(epoch uint32) uint64 {
	if stub.PersisterCalled != nil {
		return stub.PersisterCalled(epoch)
	}

	return 0
}

// IncrementSnapshotPersister -
func (stub *StateStatisticsHandlerStub) IncrementSnapshotPersister(epoch uint32) {
	if stub.IncrementSnapshotPersisterCalled != nil {
		stub.IncrementSnapshotPersisterCalled(epoch)
	}
}

// SnapshotPersister -
func (stub *StateStatisticsHandlerStub) SnapshotPersister(epoch uint32) uint64 {
	if stub.SnapshotPersisterCalled != nil {
		return stub.SnapshotPersisterCalled(epoch)
	}

	return 0
}

// IncrementTrie -
func (stub *StateStatisticsHandlerStub) IncrementTrie() {
	if stub.IncrementTrieCalled != nil {
		stub.IncrementTrieCalled()
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
