package directoryhandler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const invalidPath = "\\/\\/\\/\\/"

func TestNewDirectoryReader(t *testing.T) {
	t.Parallel()

	instance := NewDirectoryReader()
	assert.NotNil(t, instance)
	assert.NotNil(t, instance)
}

func TestDirectoryReaderListFilesAsString(t *testing.T) {
	t.Parallel()

	t.Run("invalid path should error", func(t *testing.T) {
		t.Parallel()

		dirReader := NewDirectoryReader()

		filesName, err := dirReader.ListFilesAsString(invalidPath)
		assert.NotNil(t, err)
		assert.Equal(t, "*fs.PathError", fmt.Sprintf("%T", err))
		assert.Nil(t, filesName)
	})
	t.Run("empty directory should error", func(t *testing.T) {
		t.Parallel()

		dirReader := NewDirectoryReader()

		filesName, err := dirReader.ListFilesAsString(t.TempDir())
		expectedErrorString := "no file in provided directory"
		assert.Equal(t, expectedErrorString, err.Error())
		assert.Nil(t, filesName)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dirName := t.TempDir()

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
	})
}

func TestDirectoryReaderListDirectoriesAsString(t *testing.T) {
	t.Parallel()

	t.Run("invalid path should error", func(t *testing.T) {
		t.Parallel()

		dirReader := NewDirectoryReader()

		directoriesNames, err := dirReader.ListDirectoriesAsString(invalidPath)
		assert.NotNil(t, err)
		assert.Equal(t, "*fs.PathError", fmt.Sprintf("%T", err))
		assert.Nil(t, directoriesNames)
	})
	t.Run("empty directory should error", func(t *testing.T) {
		t.Parallel()

		dirReader := NewDirectoryReader()

		directoriesNames, err := dirReader.ListDirectoriesAsString(t.TempDir())
		expectedErrorString := "no sub-directory in provided directory"
		assert.Equal(t, expectedErrorString, err.Error())
		assert.Nil(t, directoriesNames)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dirName := t.TempDir()

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
	})
}

func TestDirectoryReaderListAllAsString(t *testing.T) {
	t.Parallel()

	t.Run("invalid path should error", func(t *testing.T) {
		t.Parallel()

		dirReader := NewDirectoryReader()

		allNames, err := dirReader.ListAllAsString(invalidPath)
		assert.NotNil(t, err)
		assert.Equal(t, "*fs.PathError", fmt.Sprintf("%T", err))
		assert.Nil(t, allNames)
	})
	t.Run("empty directory should error", func(t *testing.T) {
		t.Parallel()

		dirReader := NewDirectoryReader()

		allNames, err := dirReader.ListAllAsString(t.TempDir())
		expectedErrorString := "no file or directory in provided directory"
		assert.Equal(t, expectedErrorString, err.Error())
		assert.Nil(t, allNames)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dirName := t.TempDir()

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
	})
}

func TestDirectoryReader_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var dr *directoryReader
	require.True(t, dr.IsInterfaceNil())

	dr = NewDirectoryReader()
	require.False(t, dr.IsInterfaceNil())
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
