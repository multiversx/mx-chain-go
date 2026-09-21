package syncer_test

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
)

func TestNewValidatorAccountsSyncer(t *testing.T) {
	t.Parallel()

	t.Run("invalid base args (nil hasher) should fail", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		args.Hasher = nil

		syncer, err := syncer.NewValidatorAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("invalid timeout, should fail", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		args.Timeout = 0

		s, err := syncer.NewValidatorAccountsSyncer(args)
		assert.Nil(t, s)
		assert.True(t, errors.Is(err, common.ErrInvalidTimeout))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}
		v, err := syncer.NewValidatorAccountsSyncer(args)
		require.Nil(t, err)
		require.NotNil(t, v)
	})
}

func TestValidatorAccountsSyncer_SyncAccounts(t *testing.T) {
	t.Parallel()

	key := []byte("rootHash")

	t.Run("nil storage marker", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}

		v, err := syncer.NewValidatorAccountsSyncer(args)
		require.Nil(t, err)
		require.NotNil(t, v)

		err = v.SyncAccounts(key, nil)
		require.Equal(t, syncer.ErrNilStorageMarker, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := syncer.ArgsNewValidatorAccountsSyncer{
			ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		}

		serializedLeafNode := getSerializedTrieNode(key, args.Marshalizer, args.Hasher)
		itn, err := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
		require.Nil(t, err)

		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetCalled: func(b []byte) ([]byte, error) {
				return serializedLeafNode, nil
			},
		}

		cacher := cache.NewCacherMock()
		cacher.Put(key, itn, 0)
		args.Cacher = cacher

		v, err := syncer.NewValidatorAccountsSyncer(args)
		require.Nil(t, err)

		err = v.SyncAccounts(key, storageMarker.NewDisabledStorageMarker())
		require.Nil(t, err)
	})
}

func TestValidatorAccountsSyncer_SyncAccountsWithDiskCheckRepairsMissingChild(t *testing.T) {
	for _, version := range []int{2, 3} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			testValidatorAccountsSyncerRepairsMissingChild(t, version)
		})
	}
}

func testValidatorAccountsSyncerRepairsMissingChild(t *testing.T, version int) {
	args := syncer.ArgsNewValidatorAccountsSyncer{ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs()}
	args.TrieSyncerVersion = version
	args.Timeout = 5 * time.Second
	var mut sync.Mutex
	nodes := make(map[string][]byte)
	args.TrieStorageManager = &storageManager.StorageManagerStub{
		PutCalled: func(key []byte, value []byte) error {
			mut.Lock()
			nodes[string(key)] = bytes.Clone(value)
			mut.Unlock()
			return nil
		},
		PutInEpochCalled: func(key []byte, value []byte, _ uint32) error {
			mut.Lock()
			nodes[string(key)] = bytes.Clone(value)
			mut.Unlock()
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			mut.Lock()
			defer mut.Unlock()
			value, ok := nodes[string(key)]
			if !ok {
				return nil, errors.New("node not found")
			}
			return bytes.Clone(value), nil
		},
	}
	tr, err := trie.NewTrie(args.TrieStorageManager, args.Marshalizer, args.Hasher, args.EnableEpochsHandler, 5)
	require.NoError(t, err)
	require.NoError(t, tr.Update([]byte("first"), []byte("one")))
	require.NoError(t, tr.Update([]byte("second"), []byte("two")))
	rootHash, err := tr.RootHash()
	require.NoError(t, err)
	require.NoError(t, tr.Commit())
	mut.Lock()
	_, hasRoot := nodes[string(rootHash)]
	mut.Unlock()
	require.True(t, hasRoot)
	_, err = tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash))
	require.NoError(t, err)
	hashes, err := tr.GetAllHashes()
	require.NoError(t, err)
	require.Greater(t, len(hashes), 1)

	var missingHash []byte
	var missingNode []byte
	for _, hash := range hashes {
		if bytes.Equal(hash, rootHash) {
			continue
		}
		missingHash = hash
		break
	}
	require.NotEmpty(t, missingHash)

	cacher := cache.NewCacherMock()
	args.Cacher = cacher
	requested := false
	args.RequestHandler = &testscommon.RequestHandlerStub{RequestTrieNodesForEpochCalled: func(_ uint32, requestedHashes [][]byte, _ string, epoch uint32) {
		require.Equal(t, uint32(2241), epoch)
		for _, hash := range requestedHashes {
			if !bytes.Equal(hash, missingHash) {
				continue
			}
			requested = true
			interceptedNode, createErr := trie.NewInterceptedTrieNode(missingNode, args.Hasher)
			require.NoError(t, createErr)
			cacher.Put(missingHash, interceptedNode, 0)
		}
	}}
	v, err := syncer.NewValidatorAccountsSyncer(args)
	require.NoError(t, err)
	require.NoError(t, v.SyncAccountsWithDiskCheck(rootHash, storageMarker.NewDisabledStorageMarker(), 2241))
	require.False(t, requested)
	mut.Lock()
	missingNode = bytes.Clone(nodes[string(missingHash)])
	delete(nodes, string(missingHash))
	mut.Unlock()
	require.NoError(t, v.SyncAccountsWithDiskCheck(rootHash, storageMarker.NewDisabledStorageMarker(), 2241))
	require.True(t, requested)
	mut.Lock()
	_, repaired := nodes[string(missingHash)]
	nodes[string(missingHash)] = bytes.Clone(nodes[string(rootHash)])
	mut.Unlock()
	require.True(t, repaired)
	requested = false
	require.NoError(t, v.SyncAccountsWithDiskCheck(rootHash, storageMarker.NewDisabledStorageMarker(), 2241))
	require.True(t, requested)
}

func TestValidatorAccountsSyncer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var vas *syncer.ValidatorAccountsSyncer
	assert.True(t, vas.IsInterfaceNil())

	args := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
	}
	vas, err := syncer.NewValidatorAccountsSyncer(args)
	require.Nil(t, err)
	assert.False(t, vas.IsInterfaceNil())
}
