package mock

// DirectoryReaderStub -
type DirectoryReaderStub struct {
	ListFilesAsStringCalled       func(directoryPath string) ([]string, error)
	ListDirectoriesAsStringCalled func(directoryPath string) ([]string, error)
	ListAllAsStringCalled         func(directoryPath string) ([]string, error)
}

// ListFilesAsString -
func (d *DirectoryReaderStub) ListFilesAsString(directoryPath string) ([]string, error) {
	if d.ListAllAsStringCalled != nil {
		return d.ListAllAsStringCalled(directoryPath)
	}

	return nil, nil
}

// ListDirectoriesAsString -
func (d *DirectoryReaderStub) ListDirectoriesAsString(directoryPath string) ([]string, error) {
	if d.ListDirectoriesAsStringCalled != nil {
		return d.ListDirectoriesAsStringCalled(directoryPath)
	}

	return nil, nil
}

// ListAllAsString -
func (d *DirectoryReaderStub) ListAllAsString(directoryPath string) ([]string, error) {
	if d.ListAllAsStringCalled != nil {
		return d.ListAllAsStringCalled(directoryPath)
	}

	return nil, nil
}

// IsInterfaceNil -
func (d *DirectoryReaderStub) IsInterfaceNil() bool {
	return d == nil
}
