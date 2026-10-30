package syncer

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
)

// TODO add more tests

func TestSnapshotSyncCancellationWhileWaitingForAnotherSync(t *testing.T) {
	for _, peer := range []bool{false, true} {
		base := &baseAccountsSyncer{}
		var run func(context.Context) error
		if peer {
			s := &validatorAccountsSyncer{baseAccountsSyncer: base}
			run = func(ctx context.Context) error {
				return s.SyncAccountsWithContext(ctx, []byte("root"), storageMarker.NewDisabledStorageMarker())
			}
		} else {
			s := &userAccountsSyncer{baseAccountsSyncer: base}
			run = func(ctx context.Context) error {
				return s.SyncAccountsWithContext(ctx, []byte("root"), storageMarker.NewDisabledStorageMarker())
			}
		}
		base.mutex.Lock()
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- run(ctx) }()
		cancel()
		select {
		case err := <-done:
			require.ErrorIs(t, err, core.ErrContextClosing)
		case <-time.After(time.Second):
			base.mutex.Unlock()
			t.Fatal("snapshot cancellation waited for the foreground sync lock")
		}
		base.mutex.Unlock()
	}
}

func TestUserAccountsSyncer_CancellationJoinsDataTrieWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started, release := make(chan struct{}), make(chan struct{})
	var startedOnce, releaseOnce sync.Once
	t.Cleanup(func() { cancel(); releaseOnce.Do(func() { close(release) }) })
	var admissions atomic.Int32
	args := ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		Throttler:                 &mock.ThrottlerStub{CanProcessCalled: func() bool { return admissions.Add(1) == 1 }},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
	args.Timeout = time.Minute
	args.RequestHandler = &testscommon.RequestHandlerStub{RequestTrieNodesCalled: func(_ uint32, _ [][]byte, _ string) {
		startedOnce.Do(func() { close(started) })
		<-release
	}}
	s, err := NewUserAccountsSyncer(args)
	require.NoError(t, err)
	leaves := make(chan core.KeyValueHolder, 2)
	for _, root := range []string{"missing-a", "missing-b"} {
		value, marshalErr := args.Marshalizer.Marshal(&accounts.UserAccountData{RootHash: []byte(root), Address: []byte(root)})
		require.NoError(t, marshalErr)
		leaves <- keyValStorage.NewKeyValStorage([]byte(root), value)
	}
	close(leaves)
	done := make(chan error, 1)
	go func() {
		done <- s.syncAccountDataTries(&common.TrieIteratorChannels{LeavesChan: leaves, ErrChan: errChan.NewErrChanWrapper()}, ctx, false, args.RequestHandler)
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("data trie synchronization did not start")
	}
	cancel()
	select {
	case <-done:
		t.Fatal("sync returned before its active worker finished")
	case <-time.After(20 * time.Millisecond):
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case err = <-done:
		require.ErrorIs(t, err, data.ErrTimeIsOut)
		require.True(t, args.Throttler.(*mock.ThrottlerStub).EndWasCalled())
	case <-time.After(time.Second):
		t.Fatal("data trie worker did not drain after cancellation")
	}
}

func getDefaultBaseAccSyncerArgs() ArgsNewBaseAccountsSyncer {
	return ArgsNewBaseAccountsSyncer{
		Hasher:                            &hashingMocks.HasherMock{},
		Marshalizer:                       marshallerMock.MarshalizerMock{},
		TrieStorageManager:                &storageManager.StorageManagerStub{},
		RequestHandler:                    &testscommon.RequestHandlerStub{},
		Timeout:                           time.Second,
		Cacher:                            cache.NewCacherMock(),
		UserAccountsSyncStatisticsHandler: &testscommon.SizeSyncStatisticsHandlerStub{},
		AppStatusHandler:                  &statusHandler.AppStatusHandlerStub{},
		MaxTrieLevelInMemory:              0,
		MaxHardCapForMissingNodes:         100,
		TrieSyncerVersion:                 2,
		CheckNodesOnDisk:                  false,
		EnableEpochsHandler:               &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

func TestUserAccountsSyncer_SyncAccounts(t *testing.T) {
	t.Parallel()

	args := ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		ShardId:                   0,
		Throttler:                 &mock.ThrottlerStub{},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
	syncer, err := NewUserAccountsSyncer(args)
	assert.Nil(t, err)
	assert.NotNil(t, syncer)

	err = syncer.SyncAccounts([]byte("rootHash"), nil)
	assert.Equal(t, ErrNilStorageMarker, err)
}

func TestUserAccountsSyncer_MissingDataTrieNodeFound(t *testing.T) {
	t.Parallel()

	numNodesSynced := 0
	numProcessedCalled := 0
	setNumMissingCalled := 0
	args := ArgsNewUserAccountsSyncer{
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

	tr, _ := trie.NewTrie(tsm, args.Marshalizer, args.Hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 5)
	key := []byte("key")
	value := []byte("value")
	_ = tr.Update(key, value)
	rootHash, _ := tr.RootHash()
	_ = tr.Commit()

	args.Cacher = &cache.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			interceptedNode, _ := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
			return interceptedNode, true
		},
	}

	syncer, _ := NewUserAccountsSyncer(args)
	// test that timeout watchdog is reset
	time.Sleep(args.Timeout * 2)
	syncer.MissingDataTrieNodeFound(rootHash)

	assert.Equal(t, 1, numNodesSynced)
	assert.Equal(t, 1, numProcessedCalled)
	assert.Equal(t, 1, setNumMissingCalled)
}
