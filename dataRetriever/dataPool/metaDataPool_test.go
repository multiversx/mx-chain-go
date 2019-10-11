package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewDataPool

func TestNewMetaDataPool_NilMetaBlockShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		nil,
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64SyncMapCacherStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMetaBlockPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilMiniBlockHeaderHashesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.Uint64SyncMapCacherStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMiniBlockHashesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilShardHeaderShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.Uint64SyncMapCacherStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilShardHeaderPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilHeaderNoncesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilMetaBlockNoncesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_ConfigOk(t *testing.T) {
	metaBlocks := &mock.CacherStub{}
	shardHeaders := &mock.CacherStub{}
	miniBlockheaders := &mock.ShardedDataStub{}
	hdrsNonces := &mock.Uint64SyncMapCacherStub{}

	tdp, err := dataPool.NewMetaDataPool(
		metaBlocks,
		miniBlockheaders,
		shardHeaders,
		hdrsNonces,
	)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, metaBlocks == tdp.MetaBlocks())
	assert.True(t, shardHeaders == tdp.ShardHeaders())
	assert.True(t, miniBlockheaders == tdp.MiniBlockHashes())
	assert.True(t, hdrsNonces == tdp.HeadersNonces())
}
