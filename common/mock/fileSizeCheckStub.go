package mock

type FileSizeCheckStub struct {
	GetSizeCalled func(string) (int64, error)
}

// GetSize gets the size of a file
func (fsc *FileSizeCheckStub) GetSize(path string) (int64, error) {
	if fsc.GetSizeCalled != nil {
		return fsc.GetSizeCalled(path)
	}

	return 0, nil
}

// IsInterfaceNil checks if the underlying object is nil
func (fsc *FileSizeCheckStub) IsInterfaceNil() bool {
	return fsc == nil
}
