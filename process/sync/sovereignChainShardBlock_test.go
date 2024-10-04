package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignChainShardBootstrap_ShouldErrNilShardBootstrap(t *testing.T) {
	t.Parallel()

	scsb, err := sync.NewSovereignChainShardBootstrap(nil)
	assert.Nil(t, scsb)
	assert.Equal(t, process.ErrNilShardBootstrap, err)
}

func TestNewSovereignChainShardBootstrap_ShouldWork(t *testing.T) {
	t.Parallel()

	args := CreateShardBootstrapMockArguments()
	sb, _ := sync.NewShardBootstrap(args)

	scsb, err := sync.NewSovereignChainShardBootstrap(sb)
	assert.NotNil(t, scsb)
	assert.Nil(t, err)
}

func TestNewSovereignChainShardBootstrap_SyncBlockGetValidatorNodeDBErrorShouldSync(t *testing.T) {
	t.Parallel()

	errGetNodeFromDB := core.NewGetNodeFromDBErrWithKey([]byte("key"), errors.New("get error"), dataRetriever.PeerAccountsUnit.String())
	args := createArgsForSyncBlockGetNodeDBError(errGetNodeFromDB)

	syncCalled := false
	args.ValidatorDBSyncer = &mock.AccountsDBSyncerStub{
		SyncAccountsCalled: func(rootHash []byte, _ common.StorageMarker) error {
			syncCalled = true
			return nil
		}}

	bs, _ := sync.NewShardBootstrap(args)
	scsb, _ := sync.NewSovereignChainShardBootstrap(bs)
	err := scsb.SyncBlock(context.Background())
	require.Equal(t, errGetNodeFromDB, err)
	require.True(t, syncCalled)
}
