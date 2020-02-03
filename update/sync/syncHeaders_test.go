package sync

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockHeadersSyncHandlerArgs() ArgsNewHeadersSyncHandler {
	return ArgsNewHeadersSyncHandler{
		Storage:        &mock.StorerStub{},
		Cache:          &mock.HeadersCacherStub{},
		Marshalizer:    &mock.MarshalizerFake{},
		EpochHandler:   &mock.EpochStartTriggerStub{},
		RequestHandler: &mock.RequestHandlerStub{},
	}
}

func TestHeadersSyncHandler(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSyncHandlerArgs()

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.NotNil(t, headersSyncHandler)
	require.Nil(t, err)
	require.False(t, headersSyncHandler.IsInterfaceNil())
}

func TestHeadersSyncHandler_NilStorageErr(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSyncHandlerArgs()
	args.Storage = nil

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, headersSyncHandler)
	require.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
}

func TestHeadersSyncHandler_NilCacheErr(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSyncHandlerArgs()
	args.Cache = nil

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, headersSyncHandler)
	require.Equal(t, dataRetriever.ErrNilCacher, err)
}

func TestHeadersSyncHandler_NilEpochHandlerErr(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSyncHandlerArgs()
	args.EpochHandler = nil

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, headersSyncHandler)
	require.Equal(t, dataRetriever.ErrNilEpochHandler, err)
}

func TestHeadersSyncHandler_NilMarshalizerEr(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSyncHandlerArgs()
	args.Marshalizer = nil

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, headersSyncHandler)
	require.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestHeadersSyncHandler_NilRequestHandlerEr(t *testing.T) {
	t.Parallel()

	args := createMockHeadersSyncHandlerArgs()
	args.RequestHandler = nil

	headersSyncHandler, err := NewHeadersSyncHandler(args)
	require.Nil(t, headersSyncHandler)
	require.Equal(t, process.ErrNilRequestHandler, err)
}

func TestSyncEpochStartMetaHeader_MetaBlockInStorage(t *testing.T) {
	t.Parallel()

	meta := &block.MetaBlock{}
	args := createMockHeadersSyncHandlerArgs()
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
	args := createMockHeadersSyncHandlerArgs()
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
	args := createMockHeadersSyncHandlerArgs()
	args.Cache, _ = headersCache.NewHeadersPool(config.HeadersPoolConfig{
		MaxHeadersPerShard:            1000,
		NumElementsToRemoveOnEviction: 1,
	})
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
		headersSyncHandler.metaBlockPool.AddHeader(metaHash, meta)
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
	args := createMockHeadersSyncHandlerArgs()
	args.Cache, _ = headersCache.NewHeadersPool(config.HeadersPoolConfig{
		MaxHeadersPerShard:            1000,
		NumElementsToRemoveOnEviction: 1,
	})

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
		headersSyncHandler.metaBlockPool.AddHeader(metaHash, meta)
	}()

	metaBlock, err := headersSyncHandler.SyncEpochStartMetaHeader(1, 100*time.Second)
	require.Equal(t, meta, metaBlock)
	require.Nil(t, err)

	metaBlockSync, err := headersSyncHandler.GetMetaBlock()
	require.Nil(t, err)
	require.Equal(t, meta, metaBlockSync)

}
