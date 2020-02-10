package mock

type MultiFileReaderStub struct {
	GetFileNamesCalled func() []string
	ReadNextItemCalled func(fileName string) (string, []byte, error)
	FinishCalled       func()
}

func (mf *MultiFileReaderStub) GetFileNames() []string {
	if mf.GetFileNamesCalled != nil {
		return mf.GetFileNamesCalled()
	}
	return nil
}
func (mf *MultiFileReaderStub) ReadNextItem(fileName string) (string, []byte, error) {
	if mf.ReadNextItemCalled != nil {
		return mf.ReadNextItemCalled(fileName)
	}
	return "", nil, nil
}
func (mf *MultiFileReaderStub) Finish() {
	if mf.FinishCalled != nil {
		mf.FinishCalled()
	}
}
func (mf *MultiFileReaderStub) IsInterfaceNil() bool {
	return mf == nil
}
