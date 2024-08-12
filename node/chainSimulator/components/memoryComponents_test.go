package components

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMemUnitForTries(t *testing.T) {
	t.Parallel()

	memUnitStorer := CreateMemUnitForTries()
	require.NotNil(t, memUnitStorer)

	memUnit, ok := memUnitStorer.(*trieStorage)
	require.True(t, ok)
	memUnit.SetEpochForPutOperation(0) // for coverage only
	key := []byte("key")
	data := []byte("data")
	require.NoError(t, memUnit.Put(key, data))

	require.NoError(t, memUnit.PutInEpoch(key, data, 0))
	require.NoError(t, memUnit.PutInEpochWithoutCache(key, data, 0))

	value, _, err := memUnit.GetFromOldEpochsWithoutAddingToCache(key)
	require.NoError(t, err)
	require.Equal(t, data, value)

	latest, err := memUnit.GetLatestStorageEpoch()
	require.NoError(t, err)
	require.Zero(t, latest)

	value, err = memUnit.GetFromCurrentEpoch(key)
	require.NoError(t, err)
	require.Equal(t, data, value)

	value, err = memUnit.GetFromEpoch(key, 0)
	require.NoError(t, err)
	require.Equal(t, data, value)

	value, err = memUnit.GetFromLastEpoch(key)
	require.NoError(t, err)
	require.Equal(t, data, value)

	require.NoError(t, memUnit.RemoveFromCurrentEpoch(key))
	value, err = memUnit.GetFromCurrentEpoch(key)
	require.Error(t, err)
	require.Empty(t, value)

	require.NoError(t, memUnit.PutInEpoch(key, data, 0))
	require.NoError(t, memUnit.RemoveFromAllActiveEpochs(key))
	value, err = memUnit.GetFromCurrentEpoch(key)
	require.Error(t, err)
	require.Empty(t, value)
}
