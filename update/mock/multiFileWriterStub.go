package mock

type MultiFileWriterStub struct {
	NewFileCalled func(name string) error
	WriteCalled   func(fileName string, key string, value []byte) error
	FinishCalled  func()
}

func (mfw *MultiFileWriterStub) NewFile(name string) error {
	if mfw.NewFileCalled != nil {
		return mfw.NewFileCalled(name)
	}
	return nil
}
func (mfw *MultiFileWriterStub) Write(fileName string, key string, value []byte) error {
	if mfw.WriteCalled != nil {
		return mfw.WriteCalled(fileName, key, value)
	}
	return nil
}
func (mfw *MultiFileWriterStub) Finish() {
	if mfw.FinishCalled != nil {
		mfw.FinishCalled()
	}
}
func (mfw *MultiFileWriterStub) IsInterfaceNil() bool {
	return mfw == nil
}
