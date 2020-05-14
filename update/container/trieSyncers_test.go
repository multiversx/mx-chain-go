package containers

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTrieSyncersContainer(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	require.False(t, check.IfNil(tsc))
}

func TestTrieSyncers_AddGetShouldWork(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	testKey := "key"
	testVal := &mock.TrieSyncersStub{}
	err := tsc.Add(testKey, testVal)
	require.NoError(t, err)

	res, err := tsc.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, testVal, res)
}

func TestTrieSyncers_AddMultipleShouldWork(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	testKey0 := "key0"
	testVal0 := &mock.TrieSyncersStub{}
	testKey1 := "key1"
	testVal1 := &mock.TrieSyncersStub{}

	err := tsc.AddMultiple([]string{testKey0, testKey1}, []update.TrieSyncer{testVal0, testVal1})
	require.NoError(t, err)

	res0, err := tsc.Get(testKey0)
	require.NoError(t, err)
	require.Equal(t, testVal0, res0)

	res1, err := tsc.Get(testKey1)
	require.NoError(t, err)
	require.Equal(t, testVal1, res1)

	require.Equal(t, 2, tsc.Len())
}

func TestTrieSyncers_ReplaceShouldWork(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	testKey := "key"
	testVal := &mock.TrieSyncersStub{}
	err := tsc.Add(testKey, testVal)
	require.NoError(t, err)

	res, err := tsc.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, testVal, res)

	// update
	newTestVal := &mock.TrieSyncersStub{
		StartSyncingCalled: func(_ []byte, _ context.Context) error {
			return errors.New("local err")
		},
	}
	err = tsc.Replace(testKey, newTestVal)
	require.NoError(t, err)

	res, err = tsc.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, newTestVal, res)
}

func TestTrieSyncers_DeleteShouldWork(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	testKey := "key"
	testVal := &mock.TrieSyncersStub{}
	err := tsc.Add(testKey, testVal)
	require.NoError(t, err)

	res, err := tsc.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, testVal, res)

	tsc.Remove(testKey)

	res, err = tsc.Get(testKey)
	require.Nil(t, res)
	require.Equal(t, update.ErrInvalidContainerKey, err)
}
