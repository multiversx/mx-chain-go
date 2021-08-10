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
