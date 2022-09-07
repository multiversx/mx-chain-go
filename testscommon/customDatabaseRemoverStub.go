package testscommon

// CustomDatabaseRemoverStub -
type CustomDatabaseRemoverStub struct {
	ShouldRemoveCalled func(dbIdentifier string, epoch uint32) bool
}

// ShouldRemove -
func (c *CustomDatabaseRemoverStub) ShouldRemove(dbIdentifier string, epoch uint32) bool {
	if c.ShouldRemoveCalled != nil {
		return c.ShouldRemoveCalled(dbIdentifier, epoch)
	}

	return false
}

// IsInterfaceNil -
func (c *CustomDatabaseRemoverStub) IsInterfaceNil() bool {
	return c == nil
}
