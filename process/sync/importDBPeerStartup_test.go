package sync_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

type importPeerStartupAccounts struct {
	*stateMock.AccountsStub
	initialize func() error
}

func (a *importPeerStartupAccounts) InitializeImportDBPeerState() error {
	return a.initialize()
}

func TestMetaBootstrap_ImportPeerStartupSnapshotWaitsAfterRestoration(t *testing.T) {
	args := CreateMetaBootstrapMockArguments()
	args.IsInImportMode = true
	restored := false
	args.StorageBootstrapper = &mock.StorageBootstrapperMock{
		LoadFromStorageCalled: func() error { restored = true; return nil },
	}
	started, finish := make(chan bool, 1), make(chan struct{})
	stopBeforeReplay := errors.New("stop before starting replay")
	args.ValidatorAccountsDB = &importPeerStartupAccounts{
		AccountsStub: &stateMock.AccountsStub{},
		initialize: func() error {
			started <- restored
			<-finish
			return stopBeforeReplay
		},
	}
	boot, err := sync.NewMetaBootstrap(args)
	require.NoError(t, err)
	returned := make(chan error, 1)
	go func() { returned <- boot.StartSyncingBlocks() }()
	defer close(finish)
	select {
	case wasRestored := <-started:
		require.True(t, wasRestored, "select the checkpoint before starting its peer snapshot")
	case <-time.After(time.Second):
		t.Fatal("startup snapshot was not invoked")
	}
	select {
	case err = <-returned:
		t.Fatalf("StartSyncingBlocks returned before the startup snapshot finished: %v", err)
	default:
	}
	// The initializer error stops before starting goroutines; this also checks
	// that the caller waits synchronously and propagates initialization errors.
	finish <- struct{}{}
	require.ErrorIs(t, <-returned, stopBeforeReplay)
}

func TestMetaBootstrap_ExcludedTipDoesNotStartPeerStartupSnapshot(t *testing.T) {
	args := CreateMetaBootstrapMockArguments()
	args.IsInImportMode = true
	args.StorageBootstrapper = &mock.StorageBootstrapperMock{
		LoadFromStorageCalled: func() error { return common.ErrRoundExcluded },
	}
	args.ValidatorAccountsDB = &importPeerStartupAccounts{
		AccountsStub: &stateMock.AccountsStub{},
		initialize:   func() error { t.Fatal("must not snapshot an excluded checkpoint"); return nil },
	}
	boot, err := sync.NewMetaBootstrap(args)
	require.NoError(t, err)
	require.ErrorIs(t, boot.StartSyncingBlocks(), common.ErrRoundExcluded)
}
