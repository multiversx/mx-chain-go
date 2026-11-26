package storage_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
)

func TestIsNotFoundInStorageErr_PersisterAccess(t *testing.T) {
	t.Parallel()

	cause := &os.PathError{Op: "stat", Path: "not found/epoch", Err: os.ErrPermission}
	err := fmt.Errorf("%w: %w", storage.ErrPersisterAccess, cause)
	require.False(t, storage.IsNotFoundInStorageErr(err))
	require.ErrorIs(t, err, os.ErrPermission)
	require.False(t, storage.IsNotFoundInStorageErr(nil))
	require.True(t, storage.IsNotFoundInStorageErr(storage.ErrPersisterNotFoundForEpoch))
	require.True(t, storage.IsNotFoundInStorageErr(storage.ErrKeyNotFound))
	require.True(t, storage.IsNotFoundInStorageErr(errors.New("legacy key not found")))
}
