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
			n := getLn()
			encoded, err := n.getEncodedNode(getDefaultTrieContext())
			require.NoError(t, err)
			nodeHash, err := encodeNodeAndGetHash(n, getDefaultTrieContext())
			require.NoError(t, err)
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
					require.Equal(t, nodeHash, key)
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
			intercepted, err := NewInterceptedTrieNode(encoded, getDefaultTrieContext().GetHasher())
			require.NoError(t, err)
			cacher.Put(nodeHash, intercepted, len(encoded))
			loaded, err := getNodeFromCacheOrStorage(nodeHash, cacher, getTrieContextWithCustomStorage(adapter))
			require.NoError(t, err)
			require.Equal(t, 1, reads)
			size, err := commitSyncedNode(loaded, getTrieContextWithCustomStorage(adapter))
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
	n := getLn()
	n.setDirty(false)
	_, err = commitSyncedNode(n, getTrieContextWithCustomStorage(adapter))
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
	n := getLn()
	nodeHash, err := encodeNodeAndGetHash(n, getDefaultTrieContext())
	require.NoError(t, err)
	_, err = commitSyncedNode(n, getTrieContextWithCustomStorage(adapter))
	require.NoError(t, err)
	require.Equal(t, 1, writes)
	loaded, err := getVerifiedNodeFromStorage(nodeHash, getTrieContextWithCustomStorage(adapter))
	require.NoError(t, err)
	_, err = commitSyncedNode(loaded, getTrieContextWithCustomStorage(adapter))
	require.NoError(t, err)
	require.Equal(t, 1, writes)
}

func TestGetVerifiedNodeFromStorageRejectsNonCanonicalEncoding(t *testing.T) {
	trieCtx := getDefaultTrieContext()
	t.Cleanup(func() { require.NoError(t, trieCtx.GetStorage().Close()) })
	n := getLn()
	encoded, err := n.getEncodedNode(trieCtx)
	require.NoError(t, err)
	// Unknown protobuf fields are skipped while decoding. Verification must hash
	// the decoded node's canonical encoding, as the node will be stored in that form.
	nonCanonical := append(append([]byte{}, encoded[:len(encoded)-1]...), 0x78, 0x01, encoded[len(encoded)-1])
	requestedHash := trieCtx.Compute(string(nonCanonical))
	db := &storageManager.StorageManagerStub{GetCalled: func([]byte) ([]byte, error) { return nonCanonical, nil }}
	_, err = getVerifiedNodeFromStorage(requestedHash, getTrieContextWithCustomStorage(db))
	require.ErrorIs(t, err, ErrNodeNotFound)
}
