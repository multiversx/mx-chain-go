package txsimulator

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/mock"
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

	roAccDb, err := NewReadOnlyAccountsDB(&mock.AccountsStub{})
	require.False(t, check.IfNil(roAccDb))
	require.NoError(t, err)
}

func TestReadOnlyAccountsDB_WriteOperationsShouldNotCalled(t *testing.T) {
	t.Parallel()

	failErrMsg := "this function should have not be called"
	accDb := &mock.AccountsStub{
		SaveAccountCalled: func(account state.AccountHandler) error {
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
		PruneTrieCalled: func(_ []byte, _ data.TriePruningIdentifier) {
			t.Errorf(failErrMsg)
		},
		CancelPruneCalled: func(_ []byte, _ data.TriePruningIdentifier) {
			t.Errorf(failErrMsg)
		},
		SnapshotStateCalled: func(_ []byte) {
			t.Errorf(failErrMsg)
		},
		SetStateCheckpointCalled: func(_ []byte) {
			t.Errorf(failErrMsg)
		},
		RecreateAllTriesCalled: func(_ []byte) (map[string]data.Trie, error) {
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

	roAccDb.PruneTrie(nil, data.NewRoot)

	roAccDb.CancelPrune(nil, data.NewRoot)

	roAccDb.SnapshotState(nil)

	roAccDb.SetStateCheckpoint(nil)

	_, err = roAccDb.RecreateAllTries(nil)
	require.NoError(t, err)
}

func TestReadOnlyAccountsDB_ReadOperationsShouldWork(t *testing.T) {
	t.Parallel()

	expectedAcc := &mock.AccountWrapMock{}
	expectedJournalLen := 37
	expectedRootHash := []byte("root")
	expectedAllLeaves := map[string][]byte{"1": []byte("1")}
	expectedNumCheckpoints := uint32(7)

	accDb := &mock.AccountsStub{
		GetExistingAccountCalled: func(_ []byte) (state.AccountHandler, error) {
			return expectedAcc, nil
		},
		LoadAccountCalled: func(_ []byte) (state.AccountHandler, error) {
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
		GetAllLeavesCalled: func(_ []byte) (map[string][]byte, error) {
			return expectedAllLeaves, nil
		},
		GetNumCheckpointsCalled: func() uint32 {
			return expectedNumCheckpoints
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

	actualAllLeaves, err := roAccDb.GetAllLeaves(nil)
	require.NoError(t, err)
	require.Equal(t, expectedAllLeaves, actualAllLeaves)

	actualNumCheckpoints := roAccDb.GetNumCheckpoints()
	require.Equal(t, expectedNumCheckpoints, actualNumCheckpoints)
}
