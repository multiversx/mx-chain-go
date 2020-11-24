package mock

// ReadCloserStub -
type ReadCloserStub struct {
	CloseCalled func() error
	ReadCalled  func(p []byte) (n int, err error)
}

// Read -
func (rcs *ReadCloserStub) Read(p []byte) (n int, err error) {
	if rcs.ReadCalled != nil {
		return rcs.ReadCalled(p)
	}

	return 0, nil
}

// Close -
func (rcs *ReadCloserStub) Close() error {
	if rcs.CloseCalled != nil {
		return rcs.CloseCalled()
	}

	return nil
}
