package syncer

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

// TODO add more tests

func getDefaultBaseAccSyncerArgs() ArgsNewBaseAccountsSyncer {
	return ArgsNewBaseAccountsSyncer{
		Hasher:                            &hashingMocks.HasherMock{},
		Marshalizer:                       marshallerMock.MarshalizerMock{},
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
		EnableEpochsHandler:               &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
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

func TestUserAccountsSyncer_MissingDataTrieNodeFound(t *testing.T) {
	t.Parallel()

	numNodesSynced := 0
	numProcessedCalled := 0
	setNumMissingCalled := 0
	args := ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
		ShardId:                   0,
		Throttler:                 &mock.ThrottlerStub{},
		AddressPubKeyConverter:    &testscommon.PubkeyConverterStub{},
	}
	args.TrieStorageManager = &storageManager.StorageManagerStub{
		PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
			numNodesSynced++
			return nil
		},
	}
	args.UserAccountsSyncStatisticsHandler = &testscommon.SizeSyncStatisticsHandlerStub{
		AddNumProcessedCalled: func(value int) {
			numProcessedCalled++
		},
		SetNumMissingCalled: func(rootHash []byte, value int) {
			setNumMissingCalled++
			assert.Equal(t, 0, value)
		},
	}

	var serializedLeafNode []byte
	tsm := &storageManager.StorageManagerStub{
		PutCalled: func(key []byte, val []byte) error {
			serializedLeafNode = val
			return nil
		},
	}

	tr, _ := trie.NewTrie(tsm, args.Marshalizer, args.Hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 5)
	key := []byte("key")
	value := []byte("value")
	_ = tr.Update(key, value)
	rootHash, _ := tr.RootHash()
	_ = tr.Commit()

	args.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			interceptedNode, _ := trie.NewInterceptedTrieNode(serializedLeafNode, args.Hasher)
			return interceptedNode, true
		},
	}

	syncer, _ := NewUserAccountsSyncer(args)
	// test that timeout watchdog is reset
	time.Sleep(args.Timeout * 2)
	syncer.MissingDataTrieNodeFound(rootHash)

	assert.Equal(t, 1, numNodesSynced)
	assert.Equal(t, 1, numProcessedCalled)
	assert.Equal(t, 1, setNumMissingCalled)
}
