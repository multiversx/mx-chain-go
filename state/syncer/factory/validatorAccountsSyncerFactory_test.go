package factory

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/stretchr/testify/require"
)

func getArgs() syncer.ArgsNewValidatorAccountsSyncer {
	return syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                            &hashingMocks.HasherMock{},
			Marshalizer:                       marshallerMock.MarshalizerMock{},
			TrieStorageManager:                &storageManager.StorageManagerStub{},
			RequestHandler:                    &testscommon.RequestHandlerStub{},
			Timeout:                           time.Second,
			Cacher:                            testscommon.NewCacherMock(),
			UserAccountsSyncStatisticsHandler: &testscommon.SizeSyncStatisticsHandlerStub{},
			AppStatusHandler:                  &statusHandler.AppStatusHandlerStub{},
			EnableEpochsHandler:               &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			MaxTrieLevelInMemory:              5,
			MaxHardCapForMissingNodes:         100,
			TrieSyncerVersion:                 3,
			CheckNodesOnDisk:                  false,
		},
	}

}

func TestValidatorAccountsSyncerFactory_CreateValidatorAccountsSyncer(t *testing.T) {
	t.Parallel()

	args := getArgs()
	factory := NewValidatorAccountsSyncerFactory()
	require.False(t, factory.IsInterfaceNil())

	valSyncer, err := factory.CreateValidatorAccountsSyncer(args)
	require.Nil(t, err)
	require.Equal(t, "*syncer.validatorAccountsSyncer", fmt.Sprintf("%T", valSyncer))
}
