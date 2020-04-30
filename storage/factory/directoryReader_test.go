package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryReaderListFilesAsString(t *testing.T) {
	t.Parallel()

	dirName := "./testDir1"
	_ = os.Mkdir(dirName, os.ModePerm)
	//remove created files
	defer func() {
		_ = os.RemoveAll(dirName)
	}()

	file1 := "file1"
	file2 := "file2"
	dir1 := "dir1"
	_, _ = os.Create(filepath.Join(dirName, file1))
	_, _ = os.Create(filepath.Join(dirName, file2))
	_ = os.Mkdir(filepath.Join(dirName, dir1), os.ModePerm)

	dirReader := NewDirectoryReader()

	filesName, err := dirReader.ListFilesAsString(dirName)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(filesName))
	assert.True(t, contains(filesName, file1))
	assert.True(t, contains(filesName, file2))
}

func TestDirectoryReaderListDirectoriesAsString(t *testing.T) {
	t.Parallel()

	dirName := "./testDir2"
	_ = os.Mkdir(dirName, os.ModePerm)
	//remove created files
	defer func() {
		_ = os.RemoveAll(dirName)
	}()

	file1 := "file1"
	file2 := "file2"
	dir1 := "dir1"
	_, _ = os.Create(filepath.Join(dirName, file1))
	_, _ = os.Create(filepath.Join(dirName, file2))
	_ = os.Mkdir(filepath.Join(dirName, dir1), os.ModePerm)

	dirReader := NewDirectoryReader()

	directoriesNames, err := dirReader.ListDirectoriesAsString(dirName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(directoriesNames))
	assert.True(t, contains(directoriesNames, dir1))
}

func TestDirectoryReaderListAllAsString(t *testing.T) {
	t.Parallel()

	dirName := "./testDir3"
	_ = os.Mkdir(dirName, os.ModePerm)
	//remove created files
	defer func() {
		_ = os.RemoveAll(dirName)
	}()

	file1 := "file1"
	file2 := "file2"
	dir1 := "dir1"
	_, _ = os.Create(filepath.Join(dirName, file1))
	_, _ = os.Create(filepath.Join(dirName, file2))
	_ = os.Mkdir(filepath.Join(dirName, dir1), os.ModePerm)

	dirReader := NewDirectoryReader()

	allNames, err := dirReader.ListAllAsString(dirName)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allNames))
	assert.True(t, contains(allNames, file1))
	assert.True(t, contains(allNames, file2))
	assert.True(t, contains(allNames, dir1))
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
