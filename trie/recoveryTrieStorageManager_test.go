package trie

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/stretchr/testify/require"
)

func TestRecoveryStorageWrites(t *testing.T) {
	for _, scenario := range []string{"existing", "missing", "corrupt", "write error"} {
		t.Run(scenario, func(t *testing.T) {
			failure := errors.New("storage failure")
			writes, reads := 0, 0
			n := getLn(getTestMarshalizerAndHasher())
			encoded, err := n.getEncodedNode()
			require.NoError(t, err)
			require.NoError(t, n.setHash())
			db := &storageManager.StorageManagerStub{
				GetForRecoveryCalled: func(_ []byte, epoch uint32) ([]byte, error) {
					require.Equal(t, uint32(7), epoch)
					reads++
					if scenario == "existing" {
						return encoded, nil
					}
					if scenario == "corrupt" {
						return []byte("invalid"), nil
					}
					return nil, storage.ErrKeyNotFound
				},
				PutInEpochCalled: func(key, value []byte, epoch uint32) error {
					require.Equal(t, uint32(7), epoch)
					require.Equal(t, n.getHash(), key)
					require.Equal(t, encoded, value)
					writes++
					if scenario == "write error" {
						return failure
					}
					return nil
				},
			}
			adapter, err := newStorageForTrieSync(ArgTrieSyncer{DB: db, RecoveryEpoch: core.OptionalUint32{Value: 7, HasValue: true}})
			require.NoError(t, err)
			cacher := cache.NewCacherMock()
			intercepted, err := NewInterceptedTrieNode(encoded, n.getHasher())
			require.NoError(t, err)
			cacher.Put(n.getHash(), intercepted, len(encoded))
			loaded, err := getNodeFromCacheOrStorage(n.getHash(), cacher, adapter, n.marsh, n.getHasher())
			require.NoError(t, err)
			require.Equal(t, 1, reads)
			size, err := commitSyncedNode(loaded, adapter)
			require.Equal(t, len(encoded), size)
			if scenario == "write error" {
				require.ErrorIs(t, err, failure)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, 1, reads, "commit must not read storage again")
			if scenario == "existing" {
				require.Zero(t, writes)
			} else {
				require.Equal(t, 1, writes)
			}
		})
	}
}

func TestNormalSyncStillWritesBothEpochs(t *testing.T) {
	var epochs []uint32
	db := &storageManager.StorageManagerStub{
		GetLatestStorageEpochCalled: func() (uint32, error) { return 7, nil },
		PutInEpochCalled: func(_, _ []byte, epoch uint32) error {
			epochs = append(epochs, epoch)
			return nil
		},
	}
	adapter, err := newStorageForTrieSync(ArgTrieSyncer{DB: db})
	require.NoError(t, err)
	n := getLn(getTestMarshalizerAndHasher())
	n.setDirty(false)
	_, err = commitSyncedNode(n, adapter)
	require.NoError(t, err)
	require.Equal(t, []uint32{7, 6}, epochs)
}

func TestRecoveryStaticStorage(t *testing.T) {
	var value []byte
	writes := 0
	db := &storageManager.StorageManagerStub{
		GetCalled: func(_ []byte) ([]byte, error) {
			if value == nil {
				return nil, ErrKeyNotFound
			}
			return value, nil
		},
		PutCalled: func(_, data []byte) error {
			writes++
			value = data
			return nil
		},
		PutInEpochCalled: func(_, _ []byte, _ uint32) error {
			t.Fatal("static storage must not use epoch writes")
			return nil
		},
	}
	static, err := NewTrieStorageManagerWithoutSnapshot(db)
	require.NoError(t, err)
	adapter, err := newStorageForTrieSync(ArgTrieSyncer{DB: static, RecoveryEpoch: core.OptionalUint32{Value: 7, HasValue: true}})
	require.NoError(t, err)
	n := getLn(getTestMarshalizerAndHasher())
	require.NoError(t, n.setHash())
	_, err = commitSyncedNode(n, adapter)
	require.NoError(t, err)
	require.Equal(t, 1, writes)
	loaded, err := getVerifiedNodeFromStorage(n.getHash(), adapter, n.marsh, n.getHasher())
	require.NoError(t, err)
	_, err = commitSyncedNode(loaded, adapter)
	require.NoError(t, err)
	require.Equal(t, 1, writes)
}
