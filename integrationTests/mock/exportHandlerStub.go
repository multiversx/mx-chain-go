package mock

// ExportHandlerStub -
type ExportHandlerStub struct {
	ExportAllCalled func(epoch uint32) error
}

// ExportAll -
func (e *ExportHandlerStub) ExportAll(epoch uint32) error {
	if e.ExportAllCalled != nil {
		return e.ExportAllCalled(epoch)
	}
	return nil
}

// IsInterfaceNil -
func (e *ExportHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
