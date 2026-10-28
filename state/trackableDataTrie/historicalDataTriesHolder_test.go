package trackableDataTrie_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestTrackableDataTrie_HistoricalCacheValidatesRoots(t *testing.T) {
	for _, scenario := range []string{"matching root", "different root", "root error", "nil trie", "typed nil trie"} {
		t.Run(scenario, func(t *testing.T) {
			root := []byte("account data root")
			rootChecks, loads, puts := 0, 0, 0
			cached := &trieMock.TrieStub{RootCalled: func() ([]byte, error) {
				rootChecks++
				if scenario == "root error" {
					return nil, errors.New("cannot hash cached trie")
				}
				if scenario == "different root" {
					return []byte("other account version"), nil
				}
				return root, nil
			}}
			recreated := &trieMock.TrieStub{}
			args := getDefaultArgs()
			args.DataTriesHolder = &trieMock.TriesHolderStub{
				GetCalled: func(key []byte) common.Trie {
					require.Equal(t, args.Identifier, key)
					if scenario == "nil trie" {
						return nil
					}
					if scenario == "typed nil trie" {
						var tr *trieMock.TrieStub
						return tr
					}
					return cached
				},
				PutCalled: func(key []byte, tr common.Trie) {
					puts++
					require.Equal(t, args.Identifier, key)
					require.Same(t, recreated, tr)
				},
				RemoveCalled: func([]byte) { t.Fatal("a miss must not remove another request's trie") },
			}
			tracker, err := trackableDataTrie.NewTrackableDataTrie(args)
			require.NoError(t, err)
			tracker.SetRootHash(root)
			tracker.SetDataTrieCreator(&trieMock.TrieStub{RecreateCalled: func(options common.RootHashHolder) (common.Trie, error) {
				loads++
				require.Equal(t, root, options.GetRootHash())
				return recreated, nil
			}})
			require.Zero(t, rootChecks, "binding a historical account must remain lazy")
			require.Zero(t, loads)
			expected := recreated
			if scenario == "matching root" {
				expected = cached
			}
			require.Same(t, expected, tracker.DataTrie())
			require.Same(t, expected, tracker.DataTrie(), "an attached trie must remain pinned to its account")
			if scenario == "nil trie" || scenario == "typed nil trie" {
				require.Zero(t, rootChecks)
			} else {
				require.Equal(t, 1, rootChecks)
			}
			if scenario == "matching root" {
				require.Zero(t, loads)
				require.Zero(t, puts)
			} else {
				require.Equal(t, 1, loads)
				require.Equal(t, 1, puts)
			}
		})
	}
}

func TestTrackableDataTrie_HistoricalEmptyRootRejectsCachedNonemptyTrie(t *testing.T) {
	args := getDefaultArgs()
	args.DataTriesHolder = &trieMock.TriesHolderStub{GetCalled: func([]byte) common.Trie {
		return &trieMock.TrieStub{RootCalled: func() ([]byte, error) { return []byte("other version"), nil }}
	}}
	tracker, err := trackableDataTrie.NewTrackableDataTrie(args)
	require.NoError(t, err)
	tracker.SetDataTrieCreator(&trieMock.TrieStub{RecreateCalled: func(common.RootHashHolder) (common.Trie, error) {
		t.Fatal("an account with no data root must not load another version")
		return nil, nil
	}})
	_, _, err = tracker.RetrieveValue([]byte("key"))
	require.ErrorIs(t, err, state.ErrNilTrie)
}

func TestTrackableDataTrie_HistoricalCacheMissPreservesRecreationError(t *testing.T) {
	expectedErr := errors.New("selected epoch unavailable")
	args := getDefaultArgs()
	args.DataTriesHolder = &trieMock.TriesHolderStub{
		GetCalled: func([]byte) common.Trie {
			return &trieMock.TrieStub{RootCalled: func() ([]byte, error) { return []byte("other root"), nil }}
		},
		PutCalled:    func([]byte, common.Trie) { t.Fatal("must not cache a failed recreation") },
		RemoveCalled: func([]byte) { t.Fatal("must not remove another request's trie") },
	}
	tracker, err := trackableDataTrie.NewTrackableDataTrie(args)
	require.NoError(t, err)
	tracker.SetRootHash([]byte("selected data root"))
	loads := 0
	tracker.SetDataTrieCreator(&trieMock.TrieStub{RecreateCalled: func(common.RootHashHolder) (common.Trie, error) {
		loads++
		return nil, expectedErr
	}})
	for attempt := 0; attempt < 2; attempt++ {
		_, _, err = tracker.RetrieveValue([]byte("key"))
		require.ErrorIs(t, err, expectedErr)
	}
	require.Equal(t, 2, loads)
}

func TestTrackableDataTrie_ExecutionCacheDoesNotValidateRoots(t *testing.T) {
	args := getDefaultArgs()
	cached := &trieMock.TrieStub{RootCalled: func() ([]byte, error) {
		t.Fatal("normal execution must reuse its mutable cached trie without hashing it")
		return nil, nil
	}}
	args.DataTriesHolder = &trieMock.TriesHolderStub{GetCalled: func([]byte) common.Trie { return cached }}
	args.DataTrieCreator = &trieMock.TrieStub{RecreateCalled: func(common.RootHashHolder) (common.Trie, error) {
		t.Fatal("normal execution must keep its cached trie")
		return nil, nil
	}}
	tracker, err := trackableDataTrie.NewTrackableDataTrie(args)
	require.NoError(t, err)
	tracker.SetRootHash([]byte("previously committed root"))
	require.Same(t, cached, tracker.DataTrie())
}
