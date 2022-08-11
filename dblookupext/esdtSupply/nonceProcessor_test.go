package esdtSupply

import (
	"testing"

	storageErrors "github.com/ElrondNetwork/elrond-go-storage/common/commonErrors"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNonceProcessor_shouldProcessLogs_currentNonceLowerThanProcessed(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	nonceProc := newNonceProcessor(marshalizer, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			processedBlockNonce := &ProcessedBlockNonce{
				Nonce: 10,
			}
			return marshalizer.Marshal(processedBlockNonce)
		},
	})

	shouldProcess, err := nonceProc.shouldProcessLog(9, false)
	require.Nil(t, err)
	require.False(t, shouldProcess)
}

func TestNonceProcessor_shouldProcessLogs_currentNonceHigherThanProcessed(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	nonceProc := newNonceProcessor(marshalizer, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			processedBlockNonce := &ProcessedBlockNonce{
				Nonce: 10,
			}
			return marshalizer.Marshal(processedBlockNonce)
		},
	})

	shouldProcess, err := nonceProc.shouldProcessLog(11, false)
	require.Nil(t, err)
	require.True(t, shouldProcess)
}

func TestNonceProcessor_shouldProcessLogs_nothingInStorageShouldProcess(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	nonceProc := newNonceProcessor(marshalizer, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, storageErrors.ErrKeyNotFound
		},
	})

	shouldProcess, err := nonceProc.shouldProcessLog(11, false)
	require.Nil(t, err)
	require.True(t, shouldProcess)
}

func TestNonceProcessor_shouldProcessLogs_revertNothingInStorage(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	nonceProc := newNonceProcessor(marshalizer, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, storageErrors.ErrKeyNotFound
		},
	})

	shouldProcess, err := nonceProc.shouldProcessLog(11, true)
	require.Nil(t, err)
	require.False(t, shouldProcess)
}

func TestNonceProcessor_saveNonceInStorage(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	nonceProc := newNonceProcessor(marshalizer, &storageStubs.StorerStub{
		PutCalled: func(key, data []byte) error {
			require.Equal(t, []byte(processedBlockKey), key)
			processedNonceBlock := &ProcessedBlockNonce{}
			_ = marshalizer.Unmarshal(processedNonceBlock, data)

			require.Equal(t, uint64(100), processedNonceBlock.Nonce)
			return nil
		},
	})

	err := nonceProc.saveNonceInStorage(100)
	require.Nil(t, err)
}
