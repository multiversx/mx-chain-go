package sync

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
)

type recoveryEpochProviderStub struct {
	*testscommon.CurrentEpochProviderStub
	activeForSync bool
}

func (stub *recoveryEpochProviderStub) EpochIsActiveForSync(_ uint32) bool {
	return stub.activeForSync
}

func TestRecoveryCheckpoint_ObservedEpochOnlyAffectsSyncState(t *testing.T) {
	provider := &recoveryEpochProviderStub{
		CurrentEpochProviderStub: &testscommon.CurrentEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(_ uint32) bool { return false },
		},
		activeForSync: true,
	}
	boot := &baseBootstrap{
		currentEpochProvider: provider,
		recoveryCheckpoint:   &common.RecoveryCheckpoint{},
		roundHandler:         &mock.RoundHandlerMock{},
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled:      func() data.HeaderHandler { return &block.HeaderV3{} },
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler { return &block.HeaderV3{Epoch: 2241} },
		},
		isNodeStateCalculated: true,
		isNodeSynchronized:    true,
	}
	require.Equal(t, common.NsSynchronized, boot.GetNodeState())
	provider.activeForSync = false
	require.Equal(t, common.NsNotSynchronized, boot.GetNodeState())
	boot.recoveryCheckpoint = nil
	provider.activeForSync = true
	require.Equal(t, common.NsNotSynchronized, boot.GetNodeState())
}

func TestRecoveryCheckpoint_UsesArithmeticAfterExcludedRounds(t *testing.T) {
	var currentHeader data.HeaderHandler
	arithmeticActive := false
	expectedEpoch := uint32(2241)
	provider := &recoveryEpochProviderStub{
		CurrentEpochProviderStub: &testscommon.CurrentEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				require.Equal(t, expectedEpoch, epoch)
				return arithmeticActive
			},
		},
		activeForSync: true,
	}
	boot := &baseBootstrap{
		currentEpochProvider: provider,
		recoveryCheckpoint:   &common.RecoveryCheckpoint{Round: 100, ExcludedEnd: 199},
		roundHandler:         &mock.RoundHandlerMock{},
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled:      func() data.HeaderHandler { return &block.HeaderV3{Epoch: 2241} },
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler { return currentHeader },
		},
		isNodeStateCalculated: true,
		isNodeSynchronized:    true,
	}

	require.Equal(t, common.NsSynchronized, boot.GetNodeState())
	currentHeader = &block.HeaderV3{Epoch: 2241, Round: 100}
	require.Equal(t, common.NsSynchronized, boot.GetNodeState())
	currentHeader = &block.HeaderV3{Epoch: 2241, Round: 199}
	require.Equal(t, common.NsSynchronized, boot.GetNodeState())
	currentHeader = &block.HeaderV3{Epoch: 2241, Round: 200}
	require.Equal(t, common.NsNotSynchronized, boot.GetNodeState())

	provider.activeForSync = false
	arithmeticActive = true
	currentHeader = &block.HeaderV3{Epoch: 2241, Round: 300}
	require.Equal(t, common.NsSynchronized, boot.GetNodeState())
	expectedEpoch = 2245
	currentHeader = &block.HeaderV3{Epoch: expectedEpoch, Round: 400}
	require.Equal(t, common.NsSynchronized, boot.GetNodeState())
}

func TestRecoveryCheckpoint_SyncStateRequiresCheckpoint(t *testing.T) {
	hash := bytes.Repeat([]byte{1}, common.HashSize)
	checkpoint, err := common.NewRecoveryCheckpoint(&config.Config{
		HardforkRoundExclusions: []config.HardforkRoundExclusionConfig{{StartRound: 101, EndRound: 199}},
		HardforkRecoveryCheckpoint: config.HardforkRecoveryCheckpointConfig{
			Enabled: true,
			Round:   100,
			Headers: []config.HardforkRecoveryHeaderConfig{
				{ShardID: 0, Hash: hex.EncodeToString(hash)},
				{ShardID: core.MetachainShardId, Hash: hex.EncodeToString(hash)},
			},
		},
	})
	require.NoError(t, err)

	var header data.HeaderHandler
	var currentHash []byte
	boot := &baseBootstrap{
		recoveryCheckpoint: checkpoint,
		chainHandler: &testscommon.ChainHandlerStub{GetCurrentBlockHeaderAndHashCalled: func() (data.HeaderHandler, []byte) {
			return header, currentHash
		}},
	}
	require.False(t, boot.isRecoveryCheckpointReached())

	header = &block.HeaderV3{Round: 99, ShardID: 0}
	currentHash = hash
	require.False(t, boot.isRecoveryCheckpointReached())

	header = &block.HeaderV3{Round: 100, ShardID: 0}
	currentHash = []byte("other")
	require.False(t, boot.isRecoveryCheckpointReached())
	currentHash = hash
	require.True(t, boot.isRecoveryCheckpointReached())

	header = &block.HeaderV3{Round: 200, ShardID: 0}
	require.True(t, boot.isRecoveryCheckpointReached())
}

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

type recoveryRoutingRequestHandler struct {
	process.RequestHandler
	enabled bool
	changes []bool
}

func (handler *recoveryRoutingRequestHandler) SetRecoveryTrieRequests(enabled bool) {
	handler.enabled = enabled
	handler.changes = append(handler.changes, enabled)
}

func TestLoadRecoveryCheckpointFromStorage_RecoveryRoutingLifecycle(t *testing.T) {
	for _, failStage := range []string{"none", "user", "peer", "storage", "not required"} {
		t.Run(failStage, func(t *testing.T) {
			requester := &recoveryRoutingRequestHandler{}
			expectedErr := errors.New("restore failed")
			storer := &recoveryRootsStorageBootstrapper{
				StorageBootstrapperMock: &mock.StorageBootstrapperMock{LoadFromStorageCalled: func() error {
					if failStage == "storage" {
						return expectedErr
					}
					return nil
				}},
				userRoot: []byte("user"), peerRoot: []byte("peer"), epoch: 6617, required: failStage != "not required",
			}
			boot := &baseBootstrap{
				requestHandler: requester, storageBootstrapper: storer,
				accountsDBSyncer: &mock.AccountsDBSyncerStub{SyncAccountsWithDiskCheckCalled: func(_ []byte, _ common.StorageMarker, _ uint32) error {
					require.True(t, requester.enabled)
					if failStage == "user" {
						return expectedErr
					}
					return nil
				}},
			}
			err := boot.loadRecoveryCheckpointFromStorage(func(_ []byte, _ uint32) error {
				require.True(t, requester.enabled)
				if failStage == "peer" {
					return expectedErr
				}
				return nil
			})
			if failStage == "none" || failStage == "not required" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, expectedErr)
			}
			require.False(t, requester.enabled)
			if failStage == "not required" {
				require.Empty(t, requester.changes)
			} else {
				require.Equal(t, []bool{true, false}, requester.changes)
			}
		})
	}
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

func TestLoadRecoveryCheckpointFromStorage_AllowsGenesisBeforeCheckpoint(t *testing.T) {
	storer := &recoveryRootsStorageBootstrapper{
		StorageBootstrapperMock: &mock.StorageBootstrapperMock{LoadFromStorageCalled: func() error {
			return process.ErrNotEnoughValidBlocksInStorage
		}},
	}
	boot := &baseBootstrap{
		recoveryCheckpoint:  &common.RecoveryCheckpoint{Round: 100},
		storageBootstrapper: storer,
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler { return &block.Header{Round: 0} },
		},
	}
	require.NoError(t, boot.loadRecoveryCheckpointFromStorage(nil))

	boot.chainHandler = &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler { return &block.Header{Round: 100} },
	}
	require.ErrorIs(t, boot.loadRecoveryCheckpointFromStorage(nil), process.ErrNotEnoughValidBlocksInStorage)
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
