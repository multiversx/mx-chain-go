package containers

import (
	"context"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/require"
)

var testTrieSyncersVal = &mock.TrieSyncersStub{}

func TestNewTrieSyncersContainer(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	require.False(t, check.IfNil(tsc))
}

func TestTrieSyncers_Get(t *testing.T) {
	t.Parallel()

	t.Run("missing key should error", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		val, err := tsc.Get(testKey)
		require.Equal(t, update.ErrInvalidContainerKey, err)
		require.Nil(t, val)
	})
	t.Run("invalid data should error", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()

		_ = tsc.AddInterface(testKey, "not an account db syncer")
		val, err := tsc.Get(testKey)
		require.Equal(t, update.ErrWrongTypeInContainer, err)
		require.Nil(t, val)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.Add(testKey, testTrieSyncersVal)
		require.NoError(t, err)

		res, err := tsc.Get(testKey)
		require.NoError(t, err)
		require.Equal(t, testTrieSyncersVal, res)
	})
}

func TestTrieSyncers_Add(t *testing.T) {
	t.Parallel()

	t.Run("nil value should error", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.Add(testKey, nil)
		require.Equal(t, update.ErrNilContainerElement, err)
	})
	t.Run("duplicated key should error", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.Add(testKey, testTrieSyncersVal)
		require.NoError(t, err)

		err = tsc.Add(testKey, testTrieSyncersVal)
		require.Equal(t, update.ErrContainerKeyAlreadyExists, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.Add(testKey, testTrieSyncersVal)
		require.NoError(t, err)
	})
}

func TestTrieSyncers_AddMultiple(t *testing.T) {
	t.Parallel()

	testKey0 := "key0"
	testTrieSyncersVal0 := &mock.TrieSyncersStub{}
	testKey1 := "key1"
	testTrieSyncersVal1 := &mock.TrieSyncersStub{}

	t.Run("different lengths should error", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.AddMultiple([]string{testKey0}, nil)
		require.Equal(t, update.ErrLenMismatch, err)
	})
	t.Run("duplicated keys should error on Add", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.AddMultiple([]string{testKey0, testKey1, testKey1}, []update.TrieSyncer{testTrieSyncersVal0, testTrieSyncersVal1, testTrieSyncersVal1})
		require.Equal(t, update.ErrContainerKeyAlreadyExists, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()

		err := tsc.AddMultiple([]string{testKey0, testKey1}, []update.TrieSyncer{testTrieSyncersVal0, testTrieSyncersVal1})
		require.NoError(t, err)

		res0, err := tsc.Get(testKey0)
		require.NoError(t, err)
		require.Equal(t, testTrieSyncersVal0, res0)

		res1, err := tsc.Get(testKey1)
		require.NoError(t, err)
		require.Equal(t, testTrieSyncersVal1, res1)

		require.Equal(t, 2, tsc.Len())
	})
}

func TestTrieSyncers_Replace(t *testing.T) {
	t.Parallel()

	t.Run("nil val should error", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.Replace(testKey, nil)
		require.Equal(t, update.ErrNilContainerElement, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tsc := NewTrieSyncersContainer()
		err := tsc.Add(testKey, testTrieSyncersVal)
		require.NoError(t, err)

		res, err := tsc.Get(testKey)
		require.NoError(t, err)
		require.Equal(t, testTrieSyncersVal, res)

		// update
		newtestTrieSyncersVal := &mock.TrieSyncersStub{
			StartSyncingCalled: func(_ []byte, _ context.Context) error {
				return errors.New("local err")
			},
		}
		err = tsc.Replace(testKey, newtestTrieSyncersVal)
		require.NoError(t, err)

		res, err = tsc.Get(testKey)
		require.NoError(t, err)
		require.Equal(t, newtestTrieSyncersVal, res)
	})
}

func TestTrieSyncers_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	tsc := NewTrieSyncersContainer()
	err := tsc.Add(testKey, testTrieSyncersVal)
	require.NoError(t, err)

	res, err := tsc.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, testTrieSyncersVal, res)

	tsc.Remove(testKey)

	res, err = tsc.Get(testKey)
	require.Nil(t, res)
	require.Equal(t, update.ErrInvalidContainerKey, err)
}
