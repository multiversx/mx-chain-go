package mock

// CloserStub -
type CloserStub struct {
	CloseCalled func() error
}

// Close -
func (cs *CloserStub) Close() error {
	if cs.CloseCalled != nil {
		return cs.CloseCalled()
	}

	return nil
}

// CleanerStub -
type CleanerStub struct {
	CleanCalled func()
}

// Clean -
func (cs *CleanerStub) Clean() {
	if cs.CleanCalled != nil {
		cs.CleanCalled()
	}
}
