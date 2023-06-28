package txsimulator

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/parsers"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewReadOnlyAccountsDB_NilOriginalAccountsDBShouldErr(t *testing.T) {
	t.Parallel()

	roAccDb, err := NewReadOnlyAccountsDB(nil)
	require.True(t, check.IfNil(roAccDb))
	require.Equal(t, ErrNilAccountsAdapter, err)
}

func TestNewReadOnlyAccountsDB(t *testing.T) {
	t.Parallel()

	roAccDb, err := NewReadOnlyAccountsDB(&stateMock.AccountsStub{})
	require.False(t, check.IfNil(roAccDb))
	require.NoError(t, err)
}

func TestReadOnlyAccountsDB_WriteOperationsShouldNotCalled(t *testing.T) {
	t.Parallel()

	failErrMsg := "this function should have not be called"
	accDb := &stateMock.AccountsStub{
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			t.Errorf(failErrMsg)
			return nil
		},
		RemoveAccountCalled: func(_ []byte) error {
			t.Errorf(failErrMsg)
			return nil
		},
		CommitCalled: func() ([]byte, error) {
			t.Errorf(failErrMsg)
			return nil, nil
		},
		RevertToSnapshotCalled: func(_ int) error {
			t.Errorf(failErrMsg)
			return nil
		},
		RecreateTrieCalled: func(_ []byte) error {
			t.Errorf(failErrMsg)
			return nil
		},
		PruneTrieCalled: func(_ []byte, _ state.TriePruningIdentifier, _ state.PruningHandler) {
			t.Errorf(failErrMsg)
		},
		CancelPruneCalled: func(_ []byte, _ state.TriePruningIdentifier) {
			t.Errorf(failErrMsg)
		},
		SnapshotStateCalled: func(_ []byte) {
			t.Errorf(failErrMsg)
		},
		SetStateCheckpointCalled: func(_ []byte) {
			t.Errorf(failErrMsg)
		},
		RecreateAllTriesCalled: func(_ []byte) (map[string]common.Trie, error) {
			t.Errorf(failErrMsg)
			return nil, nil
		},
	}

	roAccDb, _ := NewReadOnlyAccountsDB(accDb)
	require.NotNil(t, roAccDb)

	err := roAccDb.SaveAccount(nil)
	require.NoError(t, err)

	err = roAccDb.RemoveAccount(nil)
	require.NoError(t, err)

	_, err = roAccDb.Commit()
	require.NoError(t, err)

	err = roAccDb.RevertToSnapshot(0)
	require.NoError(t, err)

	err = roAccDb.RecreateTrie(nil)
	require.NoError(t, err)

	roAccDb.PruneTrie(nil, state.NewRoot, state.NewPruningHandler(state.EnableDataRemoval))

	roAccDb.CancelPrune(nil, state.NewRoot)

	roAccDb.SnapshotState(nil)

	roAccDb.SetStateCheckpoint(nil)

	_, err = roAccDb.RecreateAllTries(nil)
	require.NoError(t, err)
}

func TestReadOnlyAccountsDB_ReadOperationsShouldWork(t *testing.T) {
	t.Parallel()

	expectedAcc := &stateMock.AccountWrapMock{}
	expectedJournalLen := 37
	expectedRootHash := []byte("root")

	accDb := &stateMock.AccountsStub{
		GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return expectedAcc, nil
		},
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return expectedAcc, nil
		},
		JournalLenCalled: func() int {
			return expectedJournalLen
		},
		RootHashCalled: func() ([]byte, error) {
			return expectedRootHash, nil
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	roAccDb, _ := NewReadOnlyAccountsDB(accDb)
	require.NotNil(t, roAccDb)

	actualAcc, err := roAccDb.GetExistingAccount(nil)
	require.NoError(t, err)
	require.Equal(t, expectedAcc, actualAcc)

	actualAcc, err = roAccDb.LoadAccount(nil)
	require.NoError(t, err)
	require.Equal(t, expectedAcc, actualAcc)

	actualJournalLen := roAccDb.JournalLen()
	require.Equal(t, expectedJournalLen, actualJournalLen)

	actualRootHash, err := roAccDb.RootHash()
	require.NoError(t, err)
	require.Equal(t, expectedRootHash, actualRootHash)

	actualIsPruningEnabled := roAccDb.IsPruningEnabled()
	require.Equal(t, true, actualIsPruningEnabled)

	allLeaves := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = roAccDb.GetAllLeaves(allLeaves, context.Background(), nil, parsers.NewMainTrieLeafParser())
	require.NoError(t, err)

	err = allLeaves.ErrChan.ReadFromChanNonBlocking()
	require.NoError(t, err)
}
