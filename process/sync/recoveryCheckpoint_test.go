package sync

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/mock"
)

type recoveryRootsStorageBootstrapper struct {
	*mock.StorageBootstrapperMock
	userRoot []byte
	peerRoot []byte
	epoch    uint32
	required bool
}

func (stub *recoveryRootsStorageBootstrapper) RecoveryCheckpointState() ([]byte, []byte, uint32, error) {
	return stub.userRoot, stub.peerRoot, stub.epoch, nil
}

func (stub *recoveryRootsStorageBootstrapper) RecoveryCheckpointRequired() (bool, error) {
	return stub.required, nil
}

func TestLoadRecoveryCheckpointFromStorage_CompletesCheckpointTriesBeforeRestore(t *testing.T) {
	userRoot := bytes.Repeat([]byte{1}, 32)
	peerRoot := bytes.Repeat([]byte{2}, 32)
	userSynced := false
	peerSynced := false
	storer := &recoveryRootsStorageBootstrapper{
		StorageBootstrapperMock: &mock.StorageBootstrapperMock{LoadFromStorageCalled: func() error {
			require.True(t, userSynced)
			require.True(t, peerSynced)
			return nil
		}},
		userRoot: userRoot,
		peerRoot: peerRoot,
		epoch:    2241,
		required: true,
	}
	boot := &baseBootstrap{
		storageBootstrapper: storer,
		accountsDBSyncer: &mock.AccountsDBSyncerStub{SyncAccountsWithDiskCheckCalled: func(rootHash []byte, _ common.StorageMarker, epoch uint32) error {
			require.Equal(t, userRoot, rootHash)
			require.Equal(t, uint32(2241), epoch)
			userSynced = true
			return nil
		}},
	}
	err := boot.loadRecoveryCheckpointFromStorage(func(rootHash []byte, epoch uint32) error {
		require.Equal(t, peerRoot, rootHash)
		require.Equal(t, uint32(2241), epoch)
		peerSynced = true
		return nil
	})
	require.NoError(t, err)
}

func TestLoadRecoveryCheckpointFromStorage_StopsWhenCheckpointTrieCannotSync(t *testing.T) {
	syncErr := errors.New("missing trie nodes")
	loadCalled := false
	storer := &recoveryRootsStorageBootstrapper{
		StorageBootstrapperMock: &mock.StorageBootstrapperMock{LoadFromStorageCalled: func() error {
			loadCalled = true
			return nil
		}},
		userRoot: bytes.Repeat([]byte{1}, 32),
		required: true,
	}
	boot := &baseBootstrap{
		storageBootstrapper: storer,
		accountsDBSyncer: &mock.AccountsDBSyncerStub{SyncAccountsWithDiskCheckCalled: func(_ []byte, _ common.StorageMarker, _ uint32) error {
			return syncErr
		}},
	}
	err := boot.loadRecoveryCheckpointFromStorage(nil)
	require.ErrorIs(t, err, syncErr)
	require.False(t, loadCalled)
}

func TestLoadRecoveryCheckpointFromStorage_DoesNotResyncCheckpointForLaterTip(t *testing.T) {
	loadCount := 0
	storer := &recoveryRootsStorageBootstrapper{
		StorageBootstrapperMock: &mock.StorageBootstrapperMock{LoadFromStorageCalled: func() error {
			loadCount++
			return nil
		}},
	}
	boot := &baseBootstrap{
		storageBootstrapper: storer,
		accountsDBSyncer: &mock.AccountsDBSyncerStub{SyncAccountsWithDiskCheckCalled: func(_ []byte, _ common.StorageMarker, _ uint32) error {
			t.Fatal("checkpoint trie sync is not needed for a later tip")
			return nil
		}},
	}
	err := boot.loadRecoveryCheckpointFromStorage(nil)
	require.NoError(t, err)
	require.Equal(t, 1, loadCount)
}

func TestLoadRecoveryCheckpointFromStorage_DoesNotRepairLaterTipWithCheckpointState(t *testing.T) {
	missing := core.NewGetNodeFromDBErrWithKey([]byte("user"), errors.New("missing"), dataRetriever.UserAccountsUnit.String())
	loadCount := 0
	storer := &recoveryRootsStorageBootstrapper{
		StorageBootstrapperMock: &mock.StorageBootstrapperMock{LoadFromStorageCalled: func() error {
			loadCount++
			return missing
		}},
		userRoot: bytes.Repeat([]byte{1}, 32),
	}
	boot := &baseBootstrap{
		storageBootstrapper: storer,
		accountsDBSyncer: &mock.AccountsDBSyncerStub{SyncAccountsWithDiskCheckCalled: func(_ []byte, _ common.StorageMarker, _ uint32) error {
			t.Fatal("checkpoint root does not repair a later tip")
			return nil
		}},
	}
	err := boot.loadRecoveryCheckpointFromStorage(nil)
	require.Error(t, err)
	require.Equal(t, 1, loadCount)
}
