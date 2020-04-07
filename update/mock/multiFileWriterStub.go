package mock

// MultiFileWriterStub -
type MultiFileWriterStub struct {
	NewFileCalled func(name string) error
	WriteCalled   func(fileName string, key string, value []byte) error
	FinishCalled  func()
}

// NewFile -
func (mfw *MultiFileWriterStub) NewFile(name string) error {
	if mfw.NewFileCalled != nil {
		return mfw.NewFileCalled(name)
	}
	return nil
}

// Write -
func (mfw *MultiFileWriterStub) Write(fileName string, key string, value []byte) error {
	if mfw.WriteCalled != nil {
		return mfw.WriteCalled(fileName, key, value)
	}
	return nil
}

// Finish -
func (mfw *MultiFileWriterStub) Finish() {
	if mfw.FinishCalled != nil {
		mfw.FinishCalled()
	}
}

// IsInterfaceNil -
func (mfw *MultiFileWriterStub) IsInterfaceNil() bool {
	return mfw == nil
}
