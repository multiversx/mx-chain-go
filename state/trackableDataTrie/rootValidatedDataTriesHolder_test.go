package trackableDataTrie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/state/triesHolder"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/require"
)

func TestTrackableDataTrie_EnableDataTrieRootValidationKeepsCreatorAndLaziness(t *testing.T) {
	for _, matching := range []bool{false, true} {
		name := "different root"
		if matching {
			name = "matching root"
		}
		t.Run(name, func(t *testing.T) {
			root := []byte("selected data root")
			rootChecks, recreations, puts := 0, 0, 0
			cached := &trieMock.TrieStub{RootCalled: func() ([]byte, error) {
				rootChecks++
				if matching {
					return root, nil
				}
				return []byte("another state"), nil
			}}
			recreated := &trieMock.TrieStub{}
			args := getDefaultArgs()
			args.DataTrieCreator = &trieMock.TrieStub{RecreateCalled: func(options common.RootHashHolder) (common.Trie, error) {
				recreations++
				require.Equal(t, root, options.GetRootHash())
				return recreated, nil
			}}
			args.DataTriesHolder = &trieMock.TriesHolderStub{
				GetCalled: func(address []byte) common.Trie {
					require.Equal(t, args.Identifier, address)
					return cached
				},
				PutCalled: func(_ []byte, tr common.Trie) {
					puts++
					require.Same(t, recreated, tr)
				},
			}
			tracker, err := trackableDataTrie.NewTrackableDataTrie(args)
			require.NoError(t, err)
			tracker.SetRootHash(root)
			tracker.EnableDataTrieRootValidation()
			require.Zero(t, rootChecks)
			require.Zero(t, recreations)
			expected, expectedRecreations := common.Trie(cached), 0
			if !matching {
				expected, expectedRecreations = recreated, 1
			}
			for range 2 {
				dataTrie, err := tracker.DataTrie()
				require.NoError(t, err)
				require.Same(t, expected, dataTrie)
			}
			require.Equal(t, 1, rootChecks, "only initial attachment checks the cache root")
			require.Equal(t, expectedRecreations, recreations)
			require.Equal(t, expectedRecreations, puts)
		})
	}
}

func TestTrackableDataTrie_RootValidationReusesEmptyTrieWithWritableHolder(t *testing.T) {
	holder, err := triesHolder.NewDataTriesHolder(common.TenMbSize)
	require.NoError(t, err)
	args := getDefaultArgs()
	args.DataTriesHolder = holder
	recreations := 0
	args.DataTrieCreator = &trieMock.TrieStub{RecreateCalled: func(options common.RootHashHolder) (common.Trie, error) {
		require.Equal(t, common.EmptyTrieHash, options.GetRootHash())
		recreations++
		return &trieMock.TrieStub{RootCalled: func() ([]byte, error) { return common.EmptyTrieHash, nil }}, nil
	}}

	load := func() common.DataTrieHandler {
		tracker, newErr := trackableDataTrie.NewTrackableDataTrie(args)
		require.NoError(t, newErr)
		tracker.SetRootHash(common.EmptyTrieHash)
		tracker.SetDataTrieCreator(args.DataTrieCreator)
		tracker.EnableDataTrieRootValidation()
		dataTrie, loadErr := tracker.DataTrie()
		require.NoError(t, loadErr)
		return dataTrie
	}

	first := load()
	require.Same(t, first, load())
	require.Equal(t, 1, recreations)
}
