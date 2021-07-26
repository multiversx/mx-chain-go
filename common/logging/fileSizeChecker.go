package logging

import "os"

type fileSizeChecker struct {
}

// GetSize gets the size of a file
func (fsc *fileSizeChecker) GetSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// IsInterfaceNil checks if the underlying object is nil
func (fsc *fileSizeChecker) IsInterfaceNil() bool {
	return fsc == nil
}
