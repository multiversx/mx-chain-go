package files

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewMultiFileWriter_CheckArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          ArgsNewMultiFileWriter
		expectedError error
	}{
		{name: "NilImportFolder", args: ArgsNewMultiFileWriter{ExportFolder: "", ExportStore: mock.NewStorerMock()}, expectedError: update.ErrInvalidFolderName},
		{name: "NilStorer", args: ArgsNewMultiFileWriter{ExportFolder: "folder", ExportStore: nil}, expectedError: update.ErrNilStorage},
		{name: "Ok", args: ArgsNewMultiFileWriter{ExportFolder: "folder", ExportStore: mock.NewStorerMock()}, expectedError: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMultiFileWriter(tt.args)
			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestNewFile(t *testing.T) {
	t.Parallel()

	exportFolderPath := "./"
	storer := mock.NewStorerMock()

	args := ArgsNewMultiFileWriter{
		ExportFolder: exportFolderPath,
		ExportStore:  storer,
	}
	mfw, _ := NewMultiFileWriter(args)
	require.False(t, mfw.IsInterfaceNil())

	fileName := "metablock"
	err := mfw.NewFile(fileName)
	require.Nil(t, err)

	//should do nothing second time when is called with the same file name
	err = mfw.NewFile(fileName)
	require.Nil(t, err)

	if _, err := os.Stat(exportFolderPath + fileName); err != nil {
		require.Fail(t, "file wasn't created")
	}

	//remove created file
	_ = os.Remove(exportFolderPath + fileName)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	exportFolderPath := "./"
	storer := mock.NewStorerMock()

	args := ArgsNewMultiFileWriter{
		ExportFolder: exportFolderPath,
		ExportStore:  storer,
	}
	mfw, _ := NewMultiFileWriter(args)

	fileName := "fileName"
	_ = os.Remove(exportFolderPath + fileName)
	key := "dataKey"
	value := []byte("dataDataDataDataMbune")
	err := mfw.Write(fileName, key, value)
	require.Nil(t, err)

	dataFromStorer, _ := storer.Get([]byte(key))
	require.True(t, bytes.Equal(value, dataFromStorer))

	mfw.Finish()

	if _, err := os.Stat(exportFolderPath + fileName); err != nil {
		require.Fail(t, "file wasn't created")
	}

	file, _ := os.Open(exportFolderPath + fileName)
	b, _ := ioutil.ReadAll(file)
	b = b[:len(b)-1]
	require.Equal(t, []byte(hex.EncodeToString([]byte(key))), b)

	//remove created file
	_ = os.Remove(exportFolderPath + fileName)
}
