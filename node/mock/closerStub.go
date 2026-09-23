package mock

// CloserStub -
type CloserStub struct {
	CloseCalled func() error
}

// Close -
func (c *CloserStub) Close() error {
	if c.CloseCalled != nil {
		return c.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (c *CloserStub) IsInterfaceNil() bool {
	return c == nil
}
