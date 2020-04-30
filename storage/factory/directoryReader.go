package factory

import (
	"errors"
	"os"
)

type directoryReader struct {
}

// NewDirectoryReader returns a new instance of directoryReader
func NewDirectoryReader() *directoryReader {
	return &directoryReader{}
}

// ListFilesAsString will return all files' names from a directory in a slice
func (dr *directoryReader) ListFilesAsString(directoryPath string) ([]string, error) {
	_, filesNames, err := dr.listDirectoriesAndFilesAsString(directoryPath)
	if err != nil {
		return nil, err
	}

	if len(filesNames) == 0 {
		return nil, errors.New("no file in provided directory")
	}

	return filesNames, nil
}

// ListDirectoriesAsString will return all directories' names in the directory as a string slice
func (dr *directoryReader) ListDirectoriesAsString(directoryPath string) ([]string, error) {
	directoriesNames, _, err := dr.listDirectoriesAndFilesAsString(directoryPath)
	if err != nil {
		return nil, err
	}

	if len(directoriesNames) == 0 {
		return nil, errors.New("no sub-directory in provided directory")
	}

	return directoriesNames, nil
}

// ListAllAsString will return all files and directories names from the directory as a string slice
func (dr *directoryReader) ListAllAsString(directoryPath string) ([]string, error) {
	directories, files, err := dr.listDirectoriesAndFilesAsString(directoryPath)
	if err != nil {
		return nil, err
	}

	filesNames := append(directories, files...)
	if len(filesNames) == 0 {
		return nil, errors.New("no file or directory in provided directory")
	}

	return filesNames, nil
}

func (dr *directoryReader) listDirectoriesAndFilesAsString(directoryPath string) ([]string, []string, error) {
	files, err := dr.loadContent(directoryPath)
	if err != nil {
		return nil, nil, err
	}

	directoriesNames := make([]string, 0)
	filesNames := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			directoriesNames = append(directoriesNames, file.Name())
			continue
		}

		filesNames = append(filesNames, file.Name())
	}

	return directoriesNames, filesNames, nil
}

func (dr *directoryReader) loadContent(path string) ([]os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(allFiles)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	return files, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dr *directoryReader) IsInterfaceNil() bool {
	return dr == nil
}
