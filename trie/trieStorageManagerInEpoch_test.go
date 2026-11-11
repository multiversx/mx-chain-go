package trie

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// epochReadWarningsCounter counts the warnings of the epoch reads whose arguments contain the given marker. It is the
// formatter and the (discarding) writer of its log observer, both safe to be called concurrently
type epochReadWarningsCounter struct {
	marker string
	num    atomic.Int32
}

// Output counts the line
func (wc *epochReadWarningsCounter) Output(line logger.LogLineHandler) []byte {
	isWarning := line.GetLogLevel() >= int32(logger.LogWarning) && line.GetMessage() == "trieStorageManagerInEpoch"
	if isWarning && strings.Contains(strings.Join(line.GetArgs(), " "), wc.marker) {
		wc.num.Add(1)
	}

	return nil
}

// Write discards the output
func (wc *epochReadWarningsCounter) Write(p []byte) (int, error) {
	return len(p), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wc *epochReadWarningsCounter) IsInterfaceNil() bool {
	return wc == nil
}

func TestNewTrieStorageManagerInEpochNilStorageManager(t *testing.T) {
	t.Parallel()

	tsmie, err := newTrieStorageManagerInEpoch(nil, 0)
	assert.Nil(t, tsmie)
	assert.Equal(t, ErrNilTrieStorage, err)
}

func TestNewTrieStorageManagerInEpochInvalidStorageManagerType(t *testing.T) {
	t.Parallel()

	trieStorage := &storageManager.StorageManagerStub{}

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.Nil(t, tsmie)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid storage manager, type is"))
}

func TestNewTrieStorageManagerInEpochInvalidStorerType(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = testscommon.NewMemDbMock()

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.Nil(t, tsmie)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid storer, type is"))
}

func TestNewTrieStorageManagerInEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.NotNil(t, tsmie)
	assert.Nil(t, err)
}

func TestTrieStorageManagerInEpoch_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tsmie *trieStorageManagerInEpoch
	assert.True(t, tsmie.IsInterfaceNil())

	_, trieStorage := newEmptyTrie()
	tsmie, _ = newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.False(t, tsmie.IsInterfaceNil())
}

func TestTrieStorageManagerInEpoch_GetFromEpoch(t *testing.T) {
	t.Parallel()

	t.Run("epoch 0 does not panic", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromEpochCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, nil
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

		_, _ = tsmie.Get([]byte("key"))
		assert.True(t, getFromEpochCalled)
	})

	t.Run("closing error should work", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromEpochCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, storage.ErrDBIsClosed
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

		_, err := tsmie.Get([]byte("key"))
		assert.Equal(t, ErrKeyNotFound, err)
		assert.True(t, getFromEpochCalled)
	})

	t.Run("other error should work", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromEpochCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, errors.New("not closing error")
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

		_, err := tsmie.Get([]byte("key"))
		assert.Equal(t, ErrKeyNotFound, err)
		assert.True(t, getFromEpochCalled)
	})

	t.Run("a key missing in the first epoch is not a warning, another error is", func(t *testing.T) {
		t.Parallel()

		currentEpoch := uint32(5)
		expectedVal := []byte("value")
		newStorageManagerInEpoch := func(errFirstEpoch error) *trieStorageManagerInEpoch {
			_, trieStorage := newEmptyTrie()
			trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
				GetFromEpochCalled: func(_ []byte, epoch uint32) ([]byte, error) {
					if epoch == currentEpoch {
						return nil, errFirstEpoch
					}
					return expectedVal, nil
				},
			}
			tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, currentEpoch)
			return tsmie
		}
		countWarnings := func(marker string) *epochReadWarningsCounter {
			counter := &epochReadWarningsCounter{marker: marker}
			require.Nil(t, logger.AddLogObserver(counter, counter))
			t.Cleanup(func() {
				_ = logger.RemoveLogObserver(counter)
			})
			return counter
		}

		// the error of a pruning storer for a key that is not in the persister of the epoch
		missWarnings := countWarnings("missed-in-first-epoch")
		errNotFound := fmt.Errorf("key %s not found in %s", "6b6579", "missed-in-first-epoch")
		require.True(t, storage.IsNotFoundInStorageErr(errNotFound))
		val, err := newStorageManagerInEpoch(errNotFound).Get([]byte("key"))
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, val)
		assert.Zero(t, missWarnings.num.Load())

		failureWarnings := countWarnings("failed-in-first-epoch")
		val, err = newStorageManagerInEpoch(fmt.Errorf("%w: failed-in-first-epoch", storage.ErrPersisterAccess)).Get([]byte("key"))
		assert.Nil(t, err)
		assert.Equal(t, expectedVal, val)
		assert.Equal(t, int32(1), failureWarnings.num.Load())
	})

	t.Run("getFromEpoch searches more storers", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromCurrentEpochCalled := false
		getFromPreviousEpochCalled := false
		currentEpoch := uint32(5)
		expectedKey := []byte("key")
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				assert.Equal(t, expectedKey, key)
				if epoch == currentEpoch {
					getFromCurrentEpochCalled = true
				}
				if epoch == currentEpoch-1 {
					getFromPreviousEpochCalled = true
				}
				return nil, nil
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 5)

		_, _ = tsmie.Get(expectedKey)
		assert.True(t, getFromCurrentEpochCalled)
		assert.True(t, getFromPreviousEpochCalled)
	})
}
