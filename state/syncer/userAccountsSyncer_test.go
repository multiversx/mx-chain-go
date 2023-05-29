package syncer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/hashesHolder"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getDefaultUserAccountsSyncerArgs() syncer.ArgsNewUserAccountsSyncer {
	return syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		ShardId:                   1,
		Throttler:                 &mock.ThrottlerStub{},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
}

func TestNewUserAccountsSyncer(t *testing.T) {
	t.Parallel()

	t.Run("invalid base args (nil hasher) should fail", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Hasher = nil

		syncer, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("nil throttler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Throttler = nil

		syncer, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, syncer)
		assert.Equal(t, data.ErrNilThrottler, err)
	})

	t.Run("nil address pubkey converter", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.AddressPubKeyConverter = nil

		s, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, s)
		assert.Equal(t, syncer.ErrNilPubkeyConverter, err)
	})

	t.Run("invalid timeout, should fail", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Timeout = 0

		s, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, s)
		assert.True(t, errors.Is(err, common.ErrInvalidTimeout))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		syncer, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, err)
		assert.NotNil(t, syncer)
	})
}

func getSerializedTrieNode(
	key []byte,
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
) []byte {
	var serializedLeafNode []byte
	tsm := &storageManager.StorageManagerStub{
		PutCalled: func(key []byte, val []byte) error {
			serializedLeafNode = val
			return nil
		},
	}

	tr, _ := trie.NewTrie(tsm, marshaller, hasher, 5)
	_ = tr.Update(key, []byte("value"))
	_ = tr.Commit()

	return serializedLeafNode
}

func TestUserAccountsSyncer_SyncAccounts(t *testing.T) {
	t.Parallel()

	t.Run("nil storage marker", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		s, err := syncer.NewUserAccountsSyncer(args)
		assert.Nil(t, err)
		assert.NotNil(t, s)

		err = s.SyncAccounts([]byte("rootHash"), nil)
		assert.Equal(t, syncer.ErrNilStorageMarker, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Timeout = 5 * time.Second

		key := []byte("rootHash")
		serializedLeafNode := getSerializedTrieNode(key, args.Marshalizer, args.Hasher)
		itn, err := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
		require.Nil(t, err)

		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetCalled: func(b []byte) ([]byte, error) {
				return serializedLeafNode, nil
			},
		}

		cacher := testscommon.NewCacherMock()
		cacher.Put(key, itn, 0)
		args.Cacher = cacher

		s, err := syncer.NewUserAccountsSyncer(args)
		require.Nil(t, err)

		err = s.SyncAccounts(key, storageMarker.NewDisabledStorageMarker())
		require.Nil(t, err)
	})
}

func getDefaultTrieParameters() (common.StorageManager, marshal.Marshalizer, hashing.Hasher, uint) {
	marshalizer := &testscommon.ProtobufMarshalizerMock{}
	hasher := &testscommon.KeccakMock{}

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}

	args := trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.NewSnapshotPruningStorerMock(),
		CheckpointsStorer:      testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:            marshalizer,
		Hasher:                 hasher,
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
		Identifier:             "identifier",
	}

	trieStorageManager, _ := trie.NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(1)

	return trieStorageManager, args.Marshalizer, args.Hasher, maxTrieLevelInMemory
}

func emptyTrie() common.Trie {
	tr, _ := trie.NewTrie(getDefaultTrieParameters())

	return tr
}

func TestUserAccountsSyncer_SyncAccountDataTries(t *testing.T) {
	t.Parallel()

	t.Run("nil leaves chan should fail", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		s, err := syncer.NewUserAccountsSyncer(args)
		require.Nil(t, err)

		err = s.SyncAccountDataTries(nil, context.TODO())
		require.Equal(t, trie.ErrNilTrieIteratorChannels, err)
	})

	t.Run("throttler cannot process and closed context should fail", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Timeout = 5 * time.Second

		key := []byte("accRootHash")
		serializedLeafNode := getSerializedTrieNode(key, args.Marshalizer, args.Hasher)
		itn, err := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
		require.Nil(t, err)

		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetCalled: func(b []byte) ([]byte, error) {
				return serializedLeafNode, nil
			},
		}
		args.Throttler = &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return false
			},
		}

		cacher := testscommon.NewCacherMock()
		cacher.Put(key, itn, 0)
		args.Cacher = cacher

		s, err := syncer.NewUserAccountsSyncer(args)
		require.Nil(t, err)

		_, _ = trie.NewTrie(args.TrieStorageManager, args.Marshalizer, args.Hasher, 5)
		tr := emptyTrie()

		account, err := state.NewUserAccount(testscommon.TestPubKeyAlice)
		require.Nil(t, err)
		account.SetRootHash(key)

		accountBytes, err := args.Marshalizer.Marshal(account)
		require.Nil(t, err)

		_ = tr.Update([]byte("doe"), []byte("reindeer"))
		_ = tr.Update([]byte("dog"), []byte("puppy"))
		_ = tr.Update([]byte("ddog"), accountBytes)
		_ = tr.Commit()

		leavesChannels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		rootHash, err := tr.RootHash()
		require.Nil(t, err)

		err = tr.GetAllLeavesOnChannel(leavesChannels, context.TODO(), rootHash, keyBuilder.NewDisabledKeyBuilder())
		require.Nil(t, err)

		ctx, cancel := context.WithCancel(context.TODO())
		cancel()

		err = s.SyncAccountDataTries(leavesChannels, ctx)
		require.Equal(t, data.ErrTimeIsOut, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getDefaultUserAccountsSyncerArgs()
		args.Timeout = 5 * time.Second

		key := []byte("accRootHash")
		serializedLeafNode := getSerializedTrieNode(key, args.Marshalizer, args.Hasher)
		itn, err := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
		require.Nil(t, err)

		args.TrieStorageManager = &storageManager.StorageManagerStub{
			GetCalled: func(b []byte) ([]byte, error) {
				return serializedLeafNode, nil
			},
		}

		cacher := testscommon.NewCacherMock()
		cacher.Put(key, itn, 0)
		args.Cacher = cacher

		s, err := syncer.NewUserAccountsSyncer(args)
		require.Nil(t, err)

		_, _ = trie.NewTrie(args.TrieStorageManager, args.Marshalizer, args.Hasher, 5)
		tr := emptyTrie()

		account, err := state.NewUserAccount(testscommon.TestPubKeyAlice)
		require.Nil(t, err)
		account.SetRootHash(key)

		accountBytes, err := args.Marshalizer.Marshal(account)
		require.Nil(t, err)

		_ = tr.Update([]byte("doe"), []byte("reindeer"))
		_ = tr.Update([]byte("dog"), []byte("puppy"))
		_ = tr.Update([]byte("ddog"), accountBytes)
		_ = tr.Commit()

		leavesChannels := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		rootHash, err := tr.RootHash()
		require.Nil(t, err)

		err = tr.GetAllLeavesOnChannel(leavesChannels, context.TODO(), rootHash, keyBuilder.NewDisabledKeyBuilder())
		require.Nil(t, err)

		err = s.SyncAccountDataTries(leavesChannels, context.TODO())
		require.Nil(t, err)
	})
}

func TestUserAccountsSyncer_MissingDataTrieNodeFound(t *testing.T) {
	t.Parallel()

	numNodesSynced := 0
	numProcessedCalled := 0
	setNumMissingCalled := 0
	args := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		ShardId:                   0,
		Throttler:                 &mock.ThrottlerStub{},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
	args.TrieStorageManager = &storageManager.StorageManagerStub{
		PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
			numNodesSynced++
			return nil
		},
	}
	args.UserAccountsSyncStatisticsHandler = &testscommon.SizeSyncStatisticsHandlerStub{
		AddNumProcessedCalled: func(value int) {
			numProcessedCalled++
		},
		SetNumMissingCalled: func(rootHash []byte, value int) {
			setNumMissingCalled++
			assert.Equal(t, 0, value)
		},
	}

	var serializedLeafNode []byte
	tsm := &storageManager.StorageManagerStub{
		PutCalled: func(key []byte, val []byte) error {
			serializedLeafNode = val
			return nil
		},
	}

	tr, _ := trie.NewTrie(tsm, args.Marshalizer, args.Hasher, 5)
	key := []byte("key")
	value := []byte("value")
	_ = tr.Update(key, value)
	rootHash, _ := tr.RootHash()
	_ = tr.Commit()

	args.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			interceptedNode, _ := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
			return interceptedNode, true
		},
	}

	syncer, _ := syncer.NewUserAccountsSyncer(args)
	// test that timeout watchdog is reset
	time.Sleep(args.Timeout * 2)
	syncer.MissingDataTrieNodeFound(rootHash)

	assert.Equal(t, 1, numNodesSynced)
	assert.Equal(t, 1, numProcessedCalled)
	assert.Equal(t, 1, setNumMissingCalled)
}

func TestUserAccountsSyncer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var uas *syncer.UserAccountsSyncer
	assert.True(t, uas.IsInterfaceNil())

	uas, err := syncer.NewUserAccountsSyncer(getDefaultUserAccountsSyncerArgs())
	require.Nil(t, err)
	assert.False(t, uas.IsInterfaceNil())
}
