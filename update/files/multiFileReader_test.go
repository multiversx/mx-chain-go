package files

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMultiFileReader_CheckArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          ArgsNewMultiFileReader
		expectedError error
	}{
		{name: "NilImportFolder", args: ArgsNewMultiFileReader{ImportFolder: "", ImportStore: mock.NewStorerMock()}, expectedError: update.ErrInvalidFolderName},
		{name: "NilStorer", args: ArgsNewMultiFileReader{ImportFolder: "folder", ImportStore: nil}, expectedError: update.ErrNilStorage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mfr, err := NewMultiFileReader(tt.args)
			require.Nil(t, mfr)
			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestMultiFileReader_ReadFiles(t *testing.T) {
	t.Parallel()

	_ = os.Mkdir("testDir", 0755)
	defer func() {
		_ = os.RemoveAll("./testDir/")
	}()

	filePath := "./testDir"
	storer := mock.NewStorerMock()
	file1, file2 := "file1", "file2"
	key1, key2 := "key1", "key2"
	data1, data2 := []byte("data1"), []byte("data2")

	// create 2 dummy file with multi data writer
	dmw, _ := NewMultiFileWriter(ArgsNewMultiFileWriter{ExportFolder: filePath, ExportStore: storer})
	_ = dmw.Write(file1, key1, data1)
	_ = dmw.Write(file2, key2, data2)
	dmw.Finish()

	dwr, err := NewMultiFileReader(ArgsNewMultiFileReader{ImportStore: storer, ImportFolder: filePath})
	require.Nil(t, err)
	require.False(t, dwr.IsInterfaceNil())

	fileName := dwr.GetFileNames()
	require.Equal(t, file1, fileName[0])
	require.Equal(t, file2, fileName[1])

	resKey, resData, err := dwr.ReadNextItem(file1)
	require.Nil(t, err)
	require.Equal(t, key1, resKey)
	require.Equal(t, data1, resData)

	resKey, resData, err = dwr.ReadNextItem(file2)
	require.Nil(t, err)
	require.Equal(t, key2, resKey)
	require.Equal(t, data2, resData)

	dwr.Finish()
}
