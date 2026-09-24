package syncer_test

import (
	"bytes"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
	"github.com/stretchr/testify/require"
)

func TestRecoveryUserAndDataTriePersistence(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			args := getDefaultUserAccountsSyncerArgs()
			args.TrieSyncerVersion = version
			args.Timeout = 5 * time.Second
			args.Throttler, _ = throttler.NewNumGoRoutinesThrottler(2)
			var lock sync.Mutex
			nodes := make(map[string][]byte)
			writes := 0
			db := &storageManager.StorageManagerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					lock.Lock()
					defer lock.Unlock()
					value, ok := nodes[string(key)]
					if !ok {
						return nil, storage.ErrKeyNotFound
					}
					return bytes.Clone(value), nil
				},
				PutCalled: func(key, value []byte) error {
					lock.Lock()
					defer lock.Unlock()
					nodes[string(key)] = bytes.Clone(value)
					return nil
				},
				PutInEpochCalled: func(key, value []byte, epoch uint32) error {
					require.Equal(t, uint32(2241), epoch)
					lock.Lock()
					defer lock.Unlock()
					nodes[string(key)] = bytes.Clone(value)
					writes++
					return nil
				},
			}
			args.TrieStorageManager = db
			dataTrie, err := trie.NewTrie(db, args.Marshalizer, args.Hasher, args.EnableEpochsHandler, 5)
			require.NoError(t, err)
			require.NoError(t, dataTrie.Update([]byte("key"), []byte("value")))
			dataRoot, err := dataTrie.RootHash()
			require.NoError(t, err)
			require.NoError(t, dataTrie.Commit())
			mainTrie, err := trie.NewTrie(db, args.Marshalizer, args.Hasher, args.EnableEpochsHandler, 5)
			require.NoError(t, err)
			account, err := args.Marshalizer.Marshal(&accounts.UserAccountData{RootHash: dataRoot, Address: []byte("account")})
			require.NoError(t, err)
			require.NoError(t, mainTrie.Update([]byte("account"), account))
			root, err := mainTrie.RootHash()
			require.NoError(t, err)
			require.NoError(t, mainTrie.Commit())
			original := make(map[string][]byte)
			for key, value := range nodes {
				original[key] = bytes.Clone(value)
			}
			cacher := cache.NewCacherMock()
			args.Cacher = cacher
			args.RequestHandler = &testscommon.RequestHandlerStub{RequestTrieNodesForEpochCalled: func(_ uint32, hashes [][]byte, _ string, epoch uint32) {
				require.Equal(t, uint32(2241), epoch)
				for _, hash := range hashes {
					value, ok := original[string(hash)]
					require.True(t, ok)
					node, nodeErr := trie.NewInterceptedTrieNode(value, args.Hasher)
					require.NoError(t, nodeErr)
					cacher.Put(hash, node, len(value))
				}
			}}
			s, err := syncer.NewUserAccountsSyncer(args)
			require.NoError(t, err)
			marker := storageMarker.NewDisabledStorageMarker()
			require.NoError(t, s.SyncAccountsWithDiskCheck(root, marker, 2241))
			require.Zero(t, writes)
			lock.Lock()
			delete(nodes, string(dataRoot))
			delete(nodes, string(root))
			lock.Unlock()
			require.NoError(t, s.SyncAccountsWithDiskCheck(root, marker, 2241))
			require.Equal(t, 2, writes)
			require.NoError(t, s.SyncAccountsWithDiskCheck(root, marker, 2241))
			require.Equal(t, 2, writes)
		})
	}
}
