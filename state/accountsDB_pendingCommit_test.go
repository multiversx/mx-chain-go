package state_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/triesHolder"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	"github.com/stretchr/testify/require"
)

func TestAccountsDB_CommitRetainsPendingDataTriesUntilBatchSucceeds(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		t.Run(fmt.Sprintf("version-%d", version), func(t *testing.T) {
			storage, msh, hsh := stateMock.GetDefaultTrieParameters()
			t.Cleanup(func() { require.NoError(t, storage.Close()) })
			flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.FixDataTrieLeafValidationFlag)
			if version == core.AutoBalanceEnabled {
				flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
			}
			manager, err := collapseManager.NewCollapseManager(64<<20, 100)
			require.NoError(t, err)
			mainTrie, err := trie.NewTrie(storage, msh, hsh, flags, manager)
			require.NoError(t, err)
			holder, err := triesHolder.NewDataTriesHolder(1 << 20)
			require.NoError(t, err)
			args := createMockAccountsDBArgs()
			args.Trie, args.Marshaller, args.Hasher, args.DataTriesHolder = mainTrie, msh, hsh, holder
			args.AccountFactory, err = factory.NewAccountCreator(factory.ArgsAccountCreator{
				Hasher: hsh, Marshaller: msh, EnableEpochsHandler: flags,
				StateAccessesCollector: args.StateAccessesCollector, DataTriesHolder: holder, DataTrieCreator: mainTrie,
			})
			require.NoError(t, err)
			adb, err := state.NewAccountsDB(args)
			require.NoError(t, err)
			addresses := [][]byte{bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32)}
			for cycle := 0; cycle < 2; cycle++ {
				for _, address := range addresses {
					account, loadErr := adb.LoadAccount(address)
					require.NoError(t, loadErr)
					user := account.(state.UserAccountHandler)
					user.IncreaseNonce(1)
					require.NoError(t, user.SaveKeyValue([]byte("key"), []byte(fmt.Sprintf("cycle-%d/account-%d", cycle, address[0]))))
					require.NoError(t, adb.SaveAccount(user))
				}
				pendingRoot, rootErr := adb.RootHash()
				require.NoError(t, rootErr)
				attempts, failuresLeft := 0, 2
				expectedErr := errors.New("data trie commit interrupted before writing")
				originals := make([]common.Trie, len(addresses))
				for i, address := range addresses {
					originals[i] = holder.Get(address)
					require.NotNil(t, originals[i])
				}
				for _, address := range addresses {
					holder.Remove(address)
				}
				for i, address := range addresses {
					original := originals[i]
					// Retain real trie persistence and serialization. The fault is
					// injected before writing, isolating holder retention from the
					// existing node dirty-flag behavior on failed storage writes.
					holder.Put(address, &trieMock.TrieStub{
						SizeInMemoryCalled: func() int { return 700 << 10 },
						CommitCalled: func() error {
							attempts++
							if failuresLeft > 0 && attempts%2 == 0 {
								failuresLeft--
								return expectedErr
							}
							return original.Commit()
						},
					})
				}
				for attempt := 0; attempt < 2; attempt++ {
					_, err = adb.Commit()
					require.ErrorIs(t, err, expectedErr, "cycle %d attempt %d commit calls %d remaining failures %d", cycle, attempt, attempts, failuresLeft)
					_, readErr := storage.Get(pendingRoot)
					require.Error(t, readErr, "main trie must not be persisted after a failed data-trie batch")
					require.Len(t, holder.GetAll(), 2, "failed and evicted tries must remain available for retry")
				}
				root, commitErr := adb.Commit()
				require.NoError(t, commitErr)
				require.Equal(t, pendingRoot, root)
				require.Empty(t, holder.GetAll(), "successful data-trie commits must be acknowledged")
				// Recreate from persisted nodes, discarding all expanded trie and
				// account objects, then verify the actual committed state.
				require.NoError(t, adb.RecreateTrie(holders.NewDefaultRootHashesHolder(root)))
				for _, address := range addresses {
					account, loadErr := adb.GetExistingAccount(address)
					require.NoError(t, loadErr)
					user := account.(state.UserAccountHandler)
					require.Equal(t, uint64(cycle+1), user.GetNonce())
					value, _, readErr := user.RetrieveValue([]byte("key"))
					require.NoError(t, readErr)
					require.Equal(t, []byte(fmt.Sprintf("cycle-%d/account-%d", cycle, address[0])), value)
				}
			}
		})
	}
}

func TestAccountsDB_CommitAcknowledgesOnlySuccessfulDataTrieBatches(t *testing.T) {
	for _, inEpoch := range []bool{false, true} {
		for _, failure := range []string{"none", "data trie", "main trie", "empty batch"} {
			t.Run(fmt.Sprintf("inEpoch-%t/%s", inEpoch, failure), func(t *testing.T) {
				expectedErr := errors.New("commit failed")
				var events []string
				var epochs []uint32
				args := createMockAccountsDBArgs()
				args.Trie = &trieMock.TrieStub{
					CommitCalled: func() error {
						events = append(events, "main trie")
						if failure == "main trie" {
							return expectedErr
						}
						return nil
					},
					RootCalled: func() ([]byte, error) { return []byte("root"), nil },
					GetStorageManagerCalled: func() common.StorageManager {
						return &storageManager.StorageManagerStub{SetEpochForPutOperationCalled: func(epoch uint32) { epochs = append(epochs, epoch) }}
					},
				}
				var pending []common.Trie
				if failure != "empty batch" {
					for i := 0; i < 2; i++ {
						pending = append(pending, &trieMock.TrieStub{CommitCalled: func() error {
							events = append(events, fmt.Sprintf("data trie %d", i))
							if failure == "data trie" && i == 1 {
								return expectedErr
							}
							return nil
						}})
					}
				}
				args.DataTriesHolder = &trieMock.TriesHolderStub{
					GetAllCalled: func() []common.Trie { return pending },
					MarkAllAsCommittedCalled: func() {
						events = append(events, "acknowledged")
					},
				}
				adb, err := state.NewAccountsDB(args)
				require.NoError(t, err)
				if inEpoch {
					_, err = adb.CommitInEpoch(12, 11)
					require.Equal(t, []uint32{11, 12}, epochs, "restore the current write epoch on success and failure")
				} else {
					_, err = adb.Commit()
					require.Empty(t, epochs)
				}
				if failure == "data trie" || failure == "main trie" {
					require.ErrorIs(t, err, expectedErr)
				} else {
					require.NoError(t, err)
				}
				switch failure {
				case "data trie":
					require.Equal(t, []string{"data trie 0", "data trie 1"}, events)
				case "empty batch":
					require.Equal(t, []string{"acknowledged", "main trie"}, events)
				default:
					require.Equal(t, []string{"data trie 0", "data trie 1", "acknowledged", "main trie"}, events)
				}
			})
		}
	}
}
