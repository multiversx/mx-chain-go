package sync

import (
	"errors"
	"testing"
	"time"

	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/stretchr/testify/require"
)

func TestGetDataFromStorage_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	res, err := GetDataFromStorage([]byte("test"), nil)
	require.Equal(t, update.ErrNilStorage, err)
	require.Nil(t, res)
}

func TestGetDataFromStorage_NotFoundShouldErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	storer := &storageStubs.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return nil, localErr
		},
	}

	res, err := GetDataFromStorage([]byte("test"), storer)
	require.Equal(t, localErr, err)
	require.Nil(t, res)
}

func TestGetDataFromStorage_FoundShouldWork(t *testing.T) {
	t.Parallel()

	expRes := []byte("result")
	storer := &storageStubs.StorerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			return expRes, nil
		},
	}

	res, err := GetDataFromStorage([]byte("test"), storer)
	require.NoError(t, err)
	require.Equal(t, expRes, res)
}

func TestWaitFor_ShouldTimeout(t *testing.T) {
	t.Parallel()

	chanToUse := make(chan bool, 1)
	err := WaitFor(chanToUse, 10*time.Millisecond)
	require.Equal(t, update.ErrTimeIsOut, err)
}

func TestWaitFor_ShouldWorkAfterTheChannelIsWrittenIn(t *testing.T) {
	t.Parallel()

	chanToUse := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		chanToUse <- true
	}()
	err := WaitFor(chanToUse, 100*time.Millisecond)
	require.NoError(t, err)
}
