package state_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestAccountsDBApiWithHistory_ReusesEmptyCreatorUntilRootChanges(t *testing.T) {
	creatorA, creatorB := &trieMock.TrieStub{}, &trieMock.TrieStub{}
	var selectedRoot string
	var boundCreators []common.DataTrieCreator
	preparations := 0
	adapter := &stateMock.AccountsStub{
		RecreateTrieCalled: func(options common.RootHashHolder) error { selectedRoot = string(options.GetRootHash()); return nil },
		GetTrieCalled: func(root []byte) (common.Trie, error) {
			require.Empty(t, root, "context must not retain the expanded main trie")
			preparations++
			if selectedRoot == "A" {
				return creatorA, nil
			}
			return creatorB, nil
		},
		GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{SetDataTrieCreatorCalled: func(creator common.DataTrieCreator) { boundCreators = append(boundCreators, creator) }}, nil
		},
	}
	history, err := state.NewAccountsDBApiWithHistory(adapter)
	require.NoError(t, err)
	for _, root := range []string{"A", "A", "A", "B", "B"} {
		_, _, err = history.GetAccountWithBlockInfo([]byte("account"), holders.NewDefaultRootHashesHolder([]byte(root)))
		require.NoError(t, err)
	}
	require.Equal(t, 2, preparations)
	require.Len(t, boundCreators, 5)
	for _, creator := range boundCreators[:3] {
		require.Same(t, creatorA, creator)
	}
	for _, creator := range boundCreators[3:] {
		require.Same(t, creatorB, creator)
	}
}

func TestAccountsDBApiWithHistory_FailedSelectionCanBeRetried(t *testing.T) {
	expectedErr := errors.New("selection failed")
	for _, failure := range []string{"recreate", "creator error", "nil creator", "typed nil creator"} {
		t.Run(failure, func(t *testing.T) {
			var selectedRoot string
			failNext := false
			recreations, accountsReturned := 0, 0
			adapter := &stateMock.AccountsStub{
				RecreateTrieCalled: func(options common.RootHashHolder) error {
					recreations++
					if failNext && failure == "recreate" {
						return expectedErr
					}
					selectedRoot = string(options.GetRootHash())
					return nil
				},
				GetTrieCalled: func(_ []byte) (common.Trie, error) {
					if failNext {
						switch failure {
						case "creator error":
							return nil, expectedErr
						case "nil creator":
							return nil, nil
						case "typed nil creator":
							var creator *trieMock.TrieStub
							return creator, nil
						}
					}
					return &trieMock.TrieStub{}, nil
				},
				GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
					accountsReturned++
					return &stateMock.UserAccountStub{SetDataTrieCreatorCalled: func(creator common.DataTrieCreator) { require.False(t, check.IfNil(creator)) }}, nil
				},
			}
			history, err := state.NewAccountsDBApiWithHistory(adapter)
			require.NoError(t, err)
			query := func(root string) error {
				account, info, queryErr := history.GetAccountWithBlockInfo([]byte("account"), holders.NewDefaultRootHashesHolder([]byte(root)))
				if queryErr != nil {
					require.Nil(t, account)
					require.Nil(t, info)
					return queryErr
				}
				require.Equal(t, root, selectedRoot)
				require.Equal(t, []byte(root), info.GetRootHash())
				return nil
			}
			require.NoError(t, query("A"))
			for cycle := 0; cycle < 2; cycle++ {
				before := accountsReturned
				failNext = true
				err = query("B")
				if failure == "nil creator" || failure == "typed nil creator" {
					require.ErrorIs(t, err, state.ErrNilTrie)
				} else {
					require.ErrorIs(t, err, expectedErr)
				}
				require.Equal(t, before, accountsReturned, "failed selections must not return an account")
				failNext = false
				// The inner trie may already be B. A must be recreated even though it was
				// the last successfully recorded selection before the failed B request.
				beforeRecreate := recreations
				require.NoError(t, query("A"))
				require.Equal(t, beforeRecreate+1, recreations)
				require.NoError(t, query("B"))
				require.NoError(t, query("A"))
			}
		})
	}
}
