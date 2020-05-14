package mock

// MultiFileReaderStub -
type MultiFileReaderStub struct {
	GetFileNamesCalled func() []string
	ReadNextItemCalled func(fileName string) (string, []byte, error)
	FinishCalled       func()
	CloseFileCalled    func(fileName string)
}

// CloseFile -
func (mf *MultiFileReaderStub) CloseFile(fileName string) {
	if mf.CloseFileCalled != nil {
		mf.CloseFileCalled(fileName)
	}
}

// GetFileNames -
func (mf *MultiFileReaderStub) GetFileNames() []string {
	if mf.GetFileNamesCalled != nil {
		return mf.GetFileNamesCalled()
	}
	return nil
}

// ReadNextItem -
func (mf *MultiFileReaderStub) ReadNextItem(fileName string) (string, []byte, error) {
	if mf.ReadNextItemCalled != nil {
		return mf.ReadNextItemCalled(fileName)
	}
	return "", nil, nil
}

// Finish -
func (mf *MultiFileReaderStub) Finish() {
	if mf.FinishCalled != nil {
		mf.FinishCalled()
	}
}

// IsInterfaceNil -
func (mf *MultiFileReaderStub) IsInterfaceNil() bool {
	return mf == nil
}
