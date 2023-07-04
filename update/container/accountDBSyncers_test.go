package containers

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/require"
)

var (
	testKey                  = "key"
	testAccountsDBSyncersVal = &mock.AccountsDBSyncerStub{}
)

func TestNewAccountsDBSyncersContainer(t *testing.T) {
	t.Parallel()

	adsc := NewAccountsDBSyncersContainer()
	require.False(t, check.IfNil(adsc))
}

func TestAccountDBSyncers_Get(t *testing.T) {
	t.Parallel()

	t.Run("missing key should error", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		val, err := adsc.Get(testKey)
		require.Equal(t, update.ErrInvalidContainerKey, err)
		require.Nil(t, val)
	})
	t.Run("invalid data should error", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()

		_ = adsc.AddInterface(testKey, "not an account db syncer")
		val, err := adsc.Get(testKey)
		require.Equal(t, update.ErrWrongTypeInContainer, err)
		require.Nil(t, val)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()

		err := adsc.Add(testKey, testAccountsDBSyncersVal)
		require.Nil(t, err)
		val, err := adsc.Get(testKey)
		require.NoError(t, err)
		require.Equal(t, testAccountsDBSyncersVal, val)
	})
}

func TestAccountDBSyncers_Add(t *testing.T) {
	t.Parallel()

	t.Run("nil value should error", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.Add(testKey, nil)
		require.Equal(t, update.ErrNilContainerElement, err)
	})
	t.Run("duplicated key should error", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.Add(testKey, testAccountsDBSyncersVal)
		require.NoError(t, err)

		err = adsc.Add(testKey, testAccountsDBSyncersVal)
		require.Equal(t, update.ErrContainerKeyAlreadyExists, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.Add(testKey, testAccountsDBSyncersVal)
		require.NoError(t, err)
	})
}

func TestAccountDBSyncers_AddMultiple(t *testing.T) {
	t.Parallel()

	testKey0 := "key0"
	testAccountsDBSyncersVal0 := &mock.AccountsDBSyncerStub{}
	testKey1 := "key1"
	testAccountsDBSyncersVal1 := &mock.AccountsDBSyncerStub{}

	t.Run("different lengths should error", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.AddMultiple([]string{testKey0}, nil)
		require.Equal(t, update.ErrLenMismatch, err)
	})
	t.Run("duplicated keys should error on Add", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.AddMultiple([]string{testKey0, testKey1, testKey1}, []update.AccountsDBSyncer{testAccountsDBSyncersVal0, testAccountsDBSyncersVal1, testAccountsDBSyncersVal1})
		require.Equal(t, update.ErrContainerKeyAlreadyExists, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.AddMultiple([]string{testKey0, testKey1}, []update.AccountsDBSyncer{testAccountsDBSyncersVal0, testAccountsDBSyncersVal1})
		require.NoError(t, err)

		res0, err := adsc.Get(testKey0)
		require.NoError(t, err)
		require.Equal(t, testAccountsDBSyncersVal0, res0)

		res1, err := adsc.Get(testKey1)
		require.NoError(t, err)
		require.Equal(t, testAccountsDBSyncersVal1, res1)

		require.Equal(t, 2, adsc.Len())
	})
}

func TestAccountDBSyncers_Replace(t *testing.T) {
	t.Parallel()

	t.Run("nil val should error", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.Replace(testKey, nil)
		require.Equal(t, update.ErrNilContainerElement, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		adsc := NewAccountsDBSyncersContainer()
		err := adsc.Add(testKey, testAccountsDBSyncersVal)
		require.NoError(t, err)

		res, err := adsc.Get(testKey)
		require.NoError(t, err)
		require.Equal(t, testAccountsDBSyncersVal, res)

		// update
		newtestAccountsDBSyncersVal := &mock.AccountsDBSyncerStub{
			SyncAccountsCalled: func(_ []byte, _ common.StorageMarker) error {
				return errors.New("local error")
			},
		}
		err = adsc.Replace(testKey, newtestAccountsDBSyncersVal)
		require.NoError(t, err)

		res, err = adsc.Get(testKey)
		require.NoError(t, err)
		require.Equal(t, newtestAccountsDBSyncersVal, res)
	})
}

func TestAccountDBSyncers_RemoveShouldWork(t *testing.T) {
	t.Parallel()

	adsc := NewAccountsDBSyncersContainer()
	err := adsc.Add(testKey, testAccountsDBSyncersVal)
	require.NoError(t, err)

	res, err := adsc.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, testAccountsDBSyncersVal, res)

	adsc.Remove(testKey)

	res, err = adsc.Get(testKey)
	require.Nil(t, res)
	require.Equal(t, update.ErrInvalidContainerKey, err)
}
