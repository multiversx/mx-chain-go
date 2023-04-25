package syncer

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/stretchr/testify/assert"
)

// TODO add more tests

func getDefaultBaseAccSyncerArgs() ArgsNewBaseAccountsSyncer {
	return ArgsNewBaseAccountsSyncer{
		Hasher:                            &hashingMocks.HasherMock{},
		Marshalizer:                       testscommon.MarshalizerMock{},
		TrieStorageManager:                &storageManager.StorageManagerStub{},
		RequestHandler:                    &testscommon.RequestHandlerStub{},
		Timeout:                           time.Second,
		Cacher:                            testscommon.NewCacherMock(),
		UserAccountsSyncStatisticsHandler: &testscommon.SizeSyncStatisticsHandlerStub{},
		AppStatusHandler:                  &statusHandler.AppStatusHandlerStub{},
		MaxTrieLevelInMemory:              0,
		MaxHardCapForMissingNodes:         100,
		TrieSyncerVersion:                 2,
		CheckNodesOnDisk:                  false,
	}
}

func TestUserAccountsSyncer_SyncAccounts(t *testing.T) {
	t.Parallel()

	args := ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		ShardId:                   0,
		Throttler:                 &mock.ThrottlerStub{},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
	syncer, err := NewUserAccountsSyncer(args)
	assert.Nil(t, err)
	assert.NotNil(t, syncer)

	err = syncer.SyncAccounts([]byte("rootHash"), nil)
	assert.Equal(t, ErrNilStorageMarker, err)
}
