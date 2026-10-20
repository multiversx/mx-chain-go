package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/require"
)

type importPeerTestStore interface {
	storage.Storer
	GetFromCurrentEpoch([]byte) ([]byte, error)
	Rotate()
	SnapshotFinished()
}

type importPeerSnapshotStub struct {
	state.SnapshotsManager
	copy func([]byte, uint32, common.StorageManager, func()) bool
}

func (s *importPeerSnapshotStub) SnapshotPeerState(root []byte, epoch uint32, manager common.StorageManager, start func()) bool {
	return s.copy(root, epoch, manager, start)
}

func newImportPeerFixture(t *testing.T, interval uint64, copy func([]byte, uint32, common.StorageManager, func()) bool) (*state.PeerAccountsDB, common.Trie, importPeerTestStore) {
	t.Helper()
	store := pruning.NewImportDBPeerPruningStorer(7)
	msh, hsh := &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}
	manager, err := trie.NewTrieStorageManager(trie.NewTrieStorageManagerArgs{
		MainStorer: store, Marshalizer: msh, Hasher: hsh, Identifier: "peer",
		GeneralConfig: config.TrieStorageManagerConfig{SnapshotsGoroutinesPerCore: 1},
		IdleProvider:  &testscommon.ProcessStatusHandlerStub{}, StatsCollector: disabled.NewStateStatistics(),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, manager.Close()) })
	tr, err := trie.NewTrie(manager, msh, hsh, enableEpochsHandlerMock.NewEnableEpochsHandlerStub(), 1)
	require.NoError(t, err)
	argsSnapshot := getDefaultSnapshotManagerArgs()
	argsSnapshot.ProcessingMode = common.ImportDb
	argsSnapshot.ChannelsProvider = iteratorChannelsProvider.NewPeerStateIteratorChannelsProvider()
	snapshots, err := state.NewSnapshotsManager(argsSnapshot)
	require.NoError(t, err)
	require.NoError(t, snapshots.SetSyncer(&mock.AccountsDBSyncerStub{}))
	args := createMockAccountsDBArgs()
	args.Trie, args.SnapshotsManager = tr, snapshots
	if copy != nil {
		args.SnapshotsManager = &importPeerSnapshotStub{SnapshotsManager: snapshots, copy: copy}
	}
	peer, err := state.NewPeerAccountsDB(args)
	require.NoError(t, err)
	require.NoError(t, peer.EnableImportDBPeerSnapshots(store, interval))
	return peer, tr, store
}

func commitImportPeer(t *testing.T, peer *state.PeerAccountsDB, tr common.Trie, header data.HeaderHandler) []byte {
	t.Helper()
	require.NoError(t, tr.Update([]byte("changed"), []byte(fmt.Sprint(header.GetNonce()))))
	root, err := peer.Commit()
	require.NoError(t, err)
	peer.RecordImportDBPeerCommit(header, []byte(fmt.Sprint(header.GetNonce())))
	return root
}

func TestImportDBPeerSnapshotPreservesSuccessorAndRepeatsInSameEpoch(t *testing.T) {
	peer, tr, store := newImportPeerFixture(t, 5, nil)
	require.NoError(t, tr.Update([]byte("unchanged"), []byte("live")))
	commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: 10, Round: 100, Epoch: 7})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 10, Round: 100}, []byte("10"), 10, []byte("10"))
	for cycle := uint64(0); cycle < 2; cycle++ {
		nonce, round := uint64(11)+cycle*2, uint64(105)+cycle*6
		candidate := commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: nonce, Round: round, Epoch: 7})
		require.NoError(t, store.Put([]byte("obsolete"), []byte("source-only")))
		peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: nonce, Round: round}, []byte(fmt.Sprint(nonce)), nonce, []byte(fmt.Sprint(nonce)))
		_, err := store.GetFromCurrentEpoch(candidate)
		require.Error(t, err, "the round threshold must rotate before the confirming child commits")
		child := &block.MetaBlock{Nonce: nonce + 1, Round: round + 1, Epoch: 7, PrevHash: []byte(fmt.Sprint(nonce))}
		latest := commitImportPeer(t, peer, tr, child)
		peer.ProcessImportDBPeerSnapshot(child, []byte(fmt.Sprint(nonce+1)), nonce+1, []byte(fmt.Sprint(nonce+1)))
		_, err = store.Get([]byte("obsolete"))
		require.Error(t, err, "a completed snapshot must destroy the source map")
		require.False(t, peer.IsSnapshotInProgress())
		for _, root := range [][]byte{candidate, latest} {
			recreated, recreateErr := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
			require.NoError(t, recreateErr)
			value, _, getErr := recreated.Get([]byte("unchanged"))
			require.NoError(t, getErr)
			require.Equal(t, "live", string(value))
		}
		recreated, err := tr.Recreate(holders.NewDefaultRootHashesHolder(latest))
		require.NoError(t, err)
		value, _, err := recreated.Get([]byte("changed"))
		require.NoError(t, err)
		require.Equal(t, fmt.Sprint(nonce+1), string(value))
	}
}

func TestImportDBPeerSnapshotWaitsForSettledExecutionAndRetargetsAfterFiveCommits(t *testing.T) {
	var copied [][]byte
	peer, tr, store := newImportPeerFixture(t, 5, func(root []byte, _ uint32, _ common.StorageManager, start func()) bool {
		start()
		copied = append(copied, append([]byte(nil), root...))
		return true
	})
	commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 10, Round: 100, Epoch: 7})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 10, Round: 100}, []byte("10"), 0, nil)
	first := commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 11, Round: 105, Epoch: 7})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 11, Round: 105}, []byte("11"), 0, nil)
	require.NoError(t, store.Put([]byte("replacement"), []byte("keep")))
	// A duplicate imported header is not another commit for the timeout.
	for i := 0; i < 6; i++ {
		peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 11, Round: 105}, []byte("11"), 0, nil)
	}
	newRoot := commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 14, Round: 108, Epoch: 7})
	for nonce := uint64(12); nonce <= 16; nonce++ {
		peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: nonce, Round: nonce + 94, PrevHash: []byte(fmt.Sprint(nonce - 1))}, []byte(fmt.Sprint(nonce)), 0, nil)
	}
	require.Empty(t, copied, "a committed child alone does not settle its parent's execution result")
	_, err := store.Get(first)
	require.NoError(t, err, "retargeting must keep the original source")
	_, err = store.GetFromCurrentEpoch([]byte("replacement"))
	require.NoError(t, err, "retargeting must reuse the replacement")
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 17, Round: 111}, []byte("17"), 14, []byte("wrong"))
	require.Empty(t, copied, "equal settled execution nonce must match the block hash")
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 18, Round: 112}, []byte("18"), 14, []byte("14"))
	require.Equal(t, [][]byte{newRoot}, copied)
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 19, Round: 116}, []byte("19"), 14, []byte("14"))
	require.Len(t, copied, 1, "the round counter must reset on snapshot start, not initial rotation")
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 20, Round: 117}, []byte("20"), 15, []byte("15"))
	require.Len(t, copied, 2, "a later settled execution also settles the recorded earlier canonical candidate")
}

func TestImportDBPeerSnapshotDoesNotCaptureIntermediateAccountCommit(t *testing.T) {
	var copied []byte
	peer, tr, store := newImportPeerFixture(t, 5, func(root []byte, _ uint32, _ common.StorageManager, start func()) bool {
		start()
		copied = append([]byte(nil), root...)
		return true
	})
	commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: 10, Round: 100})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 10, Round: 100}, []byte("10"), 0, nil)
	require.NoError(t, store.Put([]byte("source"), []byte("present")))
	require.NoError(t, tr.Update([]byte("intermediate"), []byte("value")))
	_, err := peer.Commit() // account commit without a successful block-commit notification
	require.NoError(t, err)
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 11, Round: 105}, []byte("11"), 0, nil)
	_, err = store.GetFromCurrentEpoch([]byte("source"))
	require.NoError(t, err, "a mismatched committed root must not rotate storage")
	root := commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: 12, Round: 106})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 12, Round: 106}, []byte("12"), 0, nil)
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 13, Round: 107, PrevHash: []byte("12")}, []byte("13"), 0, nil)
	require.Equal(t, root, copied)
}

func TestImportDBPeerEpochRequestCoalescesAndFailureHasNoErrorPath(t *testing.T) {
	attempts := 0
	peer, tr, store := newImportPeerFixture(t, 100, func(_ []byte, _ uint32, _ common.StorageManager, start func()) bool {
		start()
		attempts++
		return attempts > 1
	})
	root := commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 10, Round: 100})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 10, Round: 100}, []byte("10"), 0, nil)
	require.NoError(t, store.Put([]byte("source"), []byte("present")))
	peer.SnapshotState(root, 7)
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 11, Round: 101}, []byte("11"), 0, nil)
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 12, Round: 102}, []byte("12"), 10, []byte("10"))
	require.Equal(t, 1, attempts)
	_, err := store.Get([]byte("source"))
	require.NoError(t, err, "an incomplete snapshot must not destroy the source")
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 13, Round: 103}, []byte("13"), 10, []byte("10"))
	require.Equal(t, 1, attempts, "a failed snapshot must not run again on every imported block")
	peer.SnapshotState(root, 8)
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 14, Round: 104}, []byte("14"), 10, []byte("10"))
	require.Equal(t, 2, attempts)
	_, err = store.Get([]byte("source"))
	require.Error(t, err)
}

func TestImportDBPeerSnapshotReturnWaitsForCompletion(t *testing.T) {
	args := getDefaultSnapshotManagerArgs()
	args.ProcessingMode = common.ImportDb
	args.ChannelsProvider = iteratorChannelsProvider.NewPeerStateIteratorChannelsProvider()
	sm, err := state.NewSnapshotsManager(args)
	require.NoError(t, err)
	started, unblock := make(chan struct{}), make(chan struct{})
	require.NoError(t, sm.SetSyncer(&mock.AccountsDBSyncerStub{SyncAccountsCalled: func([]byte, common.StorageMarker) error {
		close(started)
		<-unblock
		return nil
	}}))
	active := false
	manager := &storageManager.StorageManagerStub{
		TakeSnapshotCalled: func(_ string, _, _ []byte, _ *common.TrieIteratorChannels, missing chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
			missing <- []byte("missing")
			stats.SnapshotFinished()
		},
		PutInEpochWithoutCacheCalled: func(key, _ []byte, _ uint32) error {
			if string(key) == common.ActiveDBKey {
				active = true
			}
			return nil
		},
	}
	done := make(chan bool, 1)
	go func() { done <- sm.SnapshotPeerState([]byte("root"), 7, manager, func() {}) }()
	<-started
	select {
	case <-done:
		t.Fatal("snapshot returned before completion")
	default:
	}
	close(unblock)
	require.True(t, <-done)
	require.True(t, active)
	require.False(t, sm.IsSnapshotInProgress())
	// A generation with the same root and epoch must be copied again.
	manager.TakeSnapshotCalled = func(_ string, _, _ []byte, _ *common.TrieIteratorChannels, _ chan []byte, stats common.SnapshotStatisticsHandler, _ uint32) {
		stats.SnapshotFinished()
	}
	require.True(t, sm.SnapshotPeerState([]byte("root"), 7, manager, func() {}))
	manager.ShouldTakeSnapshotCalled = func() bool { return false }
	require.False(t, sm.SnapshotPeerState([]byte("skipped"), 7, manager, func() { t.Fatal("a skipped worker must not reset rounds") }))
}

func TestImportDBPeerSnapshotDoesNotHoldAccountCommitLock(t *testing.T) {
	started, unblock := make(chan struct{}), make(chan struct{})
	peer, tr, store := newImportPeerFixture(t, 5, func(_ []byte, _ uint32, _ common.StorageManager, start func()) bool {
		start()
		close(started)
		<-unblock
		return false
	})
	commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: 10, Round: 100})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 10, Round: 100}, []byte("10"), 0, nil)
	commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: 11, Round: 105})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 11, Round: 105}, []byte("11"), 0, nil)
	done := make(chan struct{})
	go func() {
		peer.ProcessImportDBPeerSnapshot(&block.MetaBlock{Nonce: 12, Round: 106, PrevHash: []byte("11")}, []byte("12"), 0, nil)
		close(done)
	}()
	<-started
	committed := make(chan []byte, 1)
	go func() {
		commitImportPeer(t, peer, tr, &block.MetaBlock{Nonce: 13, Round: 107})
		root, err := peer.RootHash()
		require.NoError(t, err)
		committed <- root
	}()
	select {
	case root := <-committed:
		_, err := store.GetFromCurrentEpoch(root)
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		close(unblock)
		<-done
		t.Fatal("snapshot blocked execution account commits")
	}
	close(unblock)
	<-done
}

func TestImportDBPeerSnapshotFinalityWinsBeforeTimeoutRetarget(t *testing.T) {
	var copied []byte
	peer, tr, _ := newImportPeerFixture(t, 5, func(root []byte, _ uint32, _ common.StorageManager, start func()) bool {
		start()
		copied = append([]byte(nil), root...)
		return true
	})
	commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 10, Round: 100})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 10, Round: 100}, []byte("10"), 0, nil)
	candidate := commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 11, Round: 105})
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 11, Round: 105}, []byte("11"), 0, nil)
	commitImportPeer(t, peer, tr, &block.MetaBlockV3{Nonce: 14, Round: 108})
	for nonce := uint64(12); nonce <= 15; nonce++ {
		for duplicate := 0; duplicate < 3; duplicate++ {
			peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: nonce, Round: nonce + 94}, []byte(fmt.Sprint(nonce)), 0, nil)
		}
	}
	require.Empty(t, copied)
	// On the fifth distinct commit, check the old candidate's finality before
	// replacing it. Duplicate notifications above must not advance its age.
	peer.ProcessImportDBPeerSnapshot(&block.MetaBlockV3{Nonce: 16, Round: 110}, []byte("16"), 11, []byte("11"))
	require.Equal(t, candidate, copied)
}
