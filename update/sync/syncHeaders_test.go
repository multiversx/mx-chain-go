package sync

import (
	"encoding/json"
	"errors"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockHeadersSuncHandlerArgs() ArgsNewHeadersSyncHandler {
	return ArgsNewHeadersSyncHandler{
		Storage: &mock.StorerStub{},
		Cache: &mock.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte)) {},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		EpochHandler:   &mock.EpochStartTriggerStub{},
		RequestHandler: &mock.RequestHandlerStub{},
	}
}

func TestHeadersSyncHandler(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSuncHandlerArgs()

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.NotNil(t, headersSyncHandler)
	require.Nil(t, err)
}

func TestSyncEpochStartMetaHeader_MetaBlockInStorage(t *testing.T) {
	t.Parallel()

	meta := &block.MetaBlock{}
	args := createMockHeadersSuncHandlerArgs()
	args.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			return json.Marshal(meta)
		},
	}

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, err)

	metaBlock, err := headersSyncHandler.SyncEpochStartMetaHeader(1, time.Second)
	require.Nil(t, err)
	require.Equal(t, meta, metaBlock)
}

func TestSyncEpochStartMetaHeader_MissingHeaderTimeout(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	args := createMockHeadersSuncHandlerArgs()
	args.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			return nil, localErr
		},
		GetFromEpochCalled: func(key []byte, epoch uint32) (bytes []byte, err error) {
			return nil, localErr
		},
	}

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, err)

	metaBlock, err := headersSyncHandler.SyncEpochStartMetaHeader(1, time.Second)
	require.Nil(t, metaBlock)
	require.Equal(t, process.ErrTimeIsOut, err)
}

func TestSyncEpochStartMetaHeader_ReceiveWrongHeaderTimeout(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	metaHash := []byte("metaHash")
	meta := &block.MetaBlock{Epoch: 1}
	args := createMockHeadersSuncHandlerArgs()
	args.Cache = mock.NewCacherMock()
	args.EpochHandler = &mock.EpochStartTriggerStub{IsEpochStartCalled: func() bool {
		return true
	}}

	args.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			return nil, localErr
		},
		GetFromEpochCalled: func(key []byte, epoch uint32) (bytes []byte, err error) {
			return nil, localErr
		},
	}

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		headersSyncHandler.metaBlockPool.Put(metaHash, meta)
	}()

	metaBlock, err := headersSyncHandler.SyncEpochStartMetaHeader(1, time.Second)
	require.Nil(t, metaBlock)
	require.Equal(t, process.ErrTimeIsOut, err)
}

func TestSyncEpochStartMetaHeader_ReceiveHeaderOk(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	metaHash := []byte("epochStartBlock_0")
	meta := &block.MetaBlock{Epoch: 1,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardId: 0, RootHash: []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{
						{Hash: []byte("hash")},
					},
				},
			},
		}}
	args := createMockHeadersSuncHandlerArgs()
	args.Cache = mock.NewCacherMock()
	args.EpochHandler = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochStartMetaHdrHashCalled: func() []byte {
			return metaHash
		},
	}

	args.Storage = &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			return nil, localErr
		},
		GetFromEpochCalled: func(key []byte, epoch uint32) (bytes []byte, err error) {
			return nil, localErr
		},
	}

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, err)

	go func() {
		time.Sleep(100 * time.Millisecond)
		headersSyncHandler.metaBlockPool.Put(metaHash, meta)
	}()

	metaBlock, err := headersSyncHandler.SyncEpochStartMetaHeader(1, 100*time.Second)
	require.Equal(t, meta, metaBlock)
	require.Nil(t, err)

	metaBlockSync, err := headersSyncHandler.GetMetaBlock()
	require.Nil(t, err)
	require.Equal(t, meta, metaBlockSync)

}
