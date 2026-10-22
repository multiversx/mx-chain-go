package node

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/stretchr/testify/require"
)

func TestTrieCollectionRealProducerReleasesPruningBuffer(t *testing.T) {
	var buffering atomic.Int32
	var mu sync.Mutex
	db := make(map[string][]byte)
	storage := &storageManager.StorageManagerStub{
		PutCalled: func(k, v []byte) error {
			mu.Lock()
			defer mu.Unlock()
			db[string(k)] = append([]byte(nil), v...)
			return nil
		},
		GetCalled: func(k []byte) ([]byte, error) {
			mu.Lock()
			defer mu.Unlock()
			v, ok := db[string(k)]
			if !ok {
				return nil, fmt.Errorf("missing node")
			}
			return v, nil
		},
		EnterPruningBufferingModeCalled: func() { buffering.Add(1) },
		ExitPruningBufferingModeCalled:  func() { buffering.Add(-1) },
	}
	tr, err := trie.NewTrie(storage, &marshal.GogoProtoMarshalizer{}, blake2b.NewBlake2b(), enableEpochsHandlerMock.NewEnableEpochsHandlerStub(), 10)
	require.NoError(t, err)
	for i := 0; i < 25; i++ {
		require.NoError(t, tr.Update([]byte(fmt.Sprint(i)), []byte("value")))
	}
	require.NoError(t, tr.Commit())
	root, err := tr.RootHash()
	require.NoError(t, err)
	source := &stateMock.StateUserAccountHandlerStub{GetAllLeavesCalled: func(ch *common.TrieIteratorChannels, ctx context.Context) error {
		return tr.GetAllLeavesOnChannel(ch, ctx, root, keyBuilder.NewKeyBuilder(), parsers.NewMainTrieLeafParser())
	}}
	n, _ := NewNode(WithTrieOperations(config.TrieOperationsConfig{MaxLeaves: 1}))
	for i := 0; i < 5; i++ {
		_, err = n.getKeys(source, context.Background())
		require.ErrorIs(t, err, ErrTrieOperationsLimit)
		require.Zero(t, buffering.Load())
	}
}

func listingStream(count int) *stateMock.StateUserAccountHandlerStub {
	return &stateMock.StateUserAccountHandlerStub{GetAllLeavesCalled: func(ch *common.TrieIteratorChannels, ctx context.Context) error {
		go func() {
			defer close(ch.LeavesChan)
			for i := 0; i < count; i++ {
				select {
				case ch.LeavesChan <- keyValStorage.NewKeyValStorage([]byte{byte(i)}, []byte("v")):
				case <-ctx.Done():
					return
				}
			}
		}()
		return nil
	}}
}

func TestTrieCollectionLimits(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cfg     config.TrieOperationsConfig
		count   int
		wantErr error
	}{
		{"exact bytes", config.TrieOperationsConfig{MaxBytes: 132}, 1, nil},
		{"over bytes", config.TrieOperationsConfig{MaxBytes: 131}, 1, ErrTrieOperationsLimit},
		{"exact leaves", config.TrieOperationsConfig{MaxLeaves: 2}, 2, nil},
		{"over leaves", config.TrieOperationsConfig{MaxLeaves: 1}, 2, ErrTrieOperationsLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n, err := NewNode(WithTrieOperations(tc.cfg))
			require.NoError(t, err)
			pairs, err := n.getKeys(listingStream(tc.count), context.Background())
			require.ErrorIs(t, err, tc.wantErr)
			if tc.wantErr != nil {
				require.Nil(t, pairs)
			} else {
				require.Len(t, pairs, tc.count)
			}
			require.Empty(t, n.trieOperationsSlots)
		})
	}
}

func TestTrieCollectionFilteredLeavesAndMetadataConsumeBudget(t *testing.T) {
	n, _ := NewNode(WithTrieOperations(config.TrieOperationsConfig{MaxLeaves: 1}))
	visited := 0
	err := n.collectTrieLeaves(context.Background(), listingStream(2), func(core.KeyValueHolder, *trieCollectionBudget) error { visited++; return nil })
	require.ErrorIs(t, err, ErrTrieOperationsLimit)
	require.Equal(t, 1, visited)
	n, _ = NewNode(WithTrieOperations(config.TrieOperationsConfig{MaxBytes: 133}))
	err = n.collectTrieLeaves(context.Background(), listingStream(1), func(_ core.KeyValueHolder, budget *trieCollectionBudget) error { return budget.chargeBytes(2) })
	require.ErrorIs(t, err, ErrTrieOperationsLimit)
}

func TestTrieCollectionFailureAndCancellation(t *testing.T) {
	expected := errors.New("storage failure")
	for _, tc := range []string{"start", "producer", "callback", "cancel"} {
		t.Run(tc, func(t *testing.T) {
			n, _ := NewNode()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			source := listingStream(2)
			visit := func(core.KeyValueHolder, *trieCollectionBudget) error { return nil }
			want := expected
			switch tc {
			case "start":
				source.GetAllLeavesCalled = func(*common.TrieIteratorChannels, context.Context) error { return expected }
			case "producer":
				source.GetAllLeavesCalled = func(ch *common.TrieIteratorChannels, _ context.Context) error {
					ch.ErrChan.WriteInChanNonBlocking(expected)
					close(ch.LeavesChan)
					return nil
				}
			case "callback":
				visit = func(core.KeyValueHolder, *trieCollectionBudget) error { return expected }
			case "cancel":
				visit = func(core.KeyValueHolder, *trieCollectionBudget) error { cancel(); return nil }
				want = ErrTrieOperationsTimeout
			}
			err := n.collectTrieLeaves(ctx, source, visit)
			require.ErrorIs(t, err, want)
			require.Empty(t, n.trieOperationsSlots)
			_, err = n.getKeys(listingStream(1), context.Background())
			require.NoError(t, err)
		})
	}
}

func TestTrieCollectionAdmissionHeldUntilProducerCloses(t *testing.T) {
	n, _ := NewNode(WithTrieOperations(config.TrieOperationsConfig{MaxConcurrentRequests: 1, MaxBytes: 1}))
	cancelled, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	source := &stateMock.StateUserAccountHandlerStub{GetAllLeavesCalled: func(ch *common.TrieIteratorChannels, ctx context.Context) error {
		go func() {
			defer close(ch.LeavesChan)
			ch.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("k"), []byte("v"))
			<-ctx.Done()
			close(cancelled)
			<-release
		}()
		return nil
	}}
	go func() { _, err := n.getKeys(source, context.Background()); done <- err }()
	select {
	case <-cancelled:
	case <-time.After(5 * time.Second):
		t.Fatal("producer was not cancelled")
	}
	_, err := n.getKeys(listingStream(1), context.Background())
	require.ErrorIs(t, err, common.ErrTrieOperationsBusy)
	close(release)
	select {
	case err = <-done:
		require.ErrorIs(t, err, ErrTrieOperationsLimit)
	case <-time.After(5 * time.Second):
		t.Fatal("scan did not exit")
	}
	require.Empty(t, n.trieOperationsSlots)
}
