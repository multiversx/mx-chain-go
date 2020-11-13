package parsers

import (
	"encoding/hex"
	"testing"

	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/stretchr/testify/require"
)

func TestStorageUpdatesParser_CreateDataFromStorageUpdate(t *testing.T) {
	t.Parallel()

	parser := NewStorageUpdatesParser()
	require.NotNil(t, parser)

	data := parser.CreateDataFromStorageUpdate(nil)
	require.Equal(t, 0, len(data))

	test := []byte("aaaa")
	stUpd := vmcommon.StorageUpdate{Offset: test, Data: test}
	stUpdates := make([]*vmcommon.StorageUpdate, 0)
	stUpdates = append(stUpdates, &stUpd, &stUpd, &stUpd)
	result := ""
	sep := "@"
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)

	data = parser.CreateDataFromStorageUpdate(stUpdates)

	require.Equal(t, result, data)
}

func TestStorageUpdatesParser_GetStorageUpdatesEmptyData(t *testing.T) {
	t.Parallel()

	parser := NewStorageUpdatesParser()
	require.NotNil(t, parser)

	stUpdates, err := parser.GetStorageUpdates("")

	require.Nil(t, stUpdates)
	require.Equal(t, ErrTokenizeFailed, err)
}

func TestStorageUpdatesParser_GetStorageUpdatesWrongData(t *testing.T) {
	t.Parallel()

	parser := NewStorageUpdatesParser()
	require.NotNil(t, parser)

	test := "test"
	result := ""
	sep := "@"
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test

	stUpdates, err := parser.GetStorageUpdates(result)

	require.Nil(t, stUpdates)
	require.Equal(t, ErrInvalidDataString, err)
}

func TestStorageUpdatesParser_GetStorageUpdates(t *testing.T) {
	t.Parallel()

	parser := NewStorageUpdatesParser()
	require.NotNil(t, parser)

	test := "aaaa"
	result := ""
	sep := "@"
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	stUpdates, err := parser.GetStorageUpdates(result)

	require.Nil(t, err)
	for i := 0; i < 2; i++ {
		require.Equal(t, test, hex.EncodeToString(stUpdates[i].Data))
		require.Equal(t, test, hex.EncodeToString(stUpdates[i].Offset))
	}
}
