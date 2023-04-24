package mock

import "sync"

// ThrottlerStub -
type ThrottlerStub struct {
	CanProcessCalled      func() bool
	StartProcessingCalled func()
	EndProcessingCalled   func()
	mutState              sync.RWMutex
	startWasCalled        bool
	endWasCalled          bool
}

// CanProcess -
func (ts *ThrottlerStub) CanProcess() bool {
	if ts.CanProcessCalled != nil {
		return ts.CanProcessCalled()
	}

	return true
}

// StartProcessing -
func (ts *ThrottlerStub) StartProcessing() {
	ts.mutState.Lock()
	ts.startWasCalled = true
	ts.mutState.Unlock()

	if ts.StartProcessingCalled != nil {
		ts.StartProcessingCalled()
	}
}

// EndProcessing -
func (ts *ThrottlerStub) EndProcessing() {
	ts.mutState.Lock()
	ts.endWasCalled = true
	ts.mutState.Unlock()

	if ts.EndProcessingCalled != nil {
		ts.EndProcessingCalled()
	}
}

// StartWasCalled -
func (ts *ThrottlerStub) StartWasCalled() bool {
	ts.mutState.RLock()
	defer ts.mutState.RUnlock()

	return ts.startWasCalled
}

// EndWasCalled -
func (ts *ThrottlerStub) EndWasCalled() bool {
	ts.mutState.RLock()
	defer ts.mutState.RUnlock()

	return ts.endWasCalled
}

// IsInterfaceNil -
func (ts *ThrottlerStub) IsInterfaceNil() bool {
	return ts == nil
}
