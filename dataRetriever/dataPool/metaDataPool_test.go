package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/dataPool"
	"github.com/stretchr/testify/assert"
)

//------- NewDataPool

func TestNewMetaDataPool_NilMetaBlockShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		nil,
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{})

	assert.Equal(t, data.ErrNilMetaBlockPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilMiniBlockHeaderHashesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.Uint64CacherStub{})

	assert.Equal(t, data.ErrNilMiniBlockHashesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilShardHeaderShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.Uint64CacherStub{})

	assert.Equal(t, data.ErrNilShardHeaderPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilMetaHeaderNouncesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		nil)

	assert.Equal(t, dataRetriever.ErrNilMetaBlockNoncesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_ConfigOk(t *testing.T) {
	metaChainBlocks := &mock.CacherStub{}
	shardHeaders := &mock.CacherStub{}
	miniBlockheaders := &mock.ShardedDataStub{}
	metaBlockNonce := &mock.Uint64CacherStub{}

	tdp, err := dataPool.NewMetaDataPool(
		metaChainBlocks,
		miniBlockheaders,
		shardHeaders,
		metaBlockNonce)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, metaChainBlocks == tdp.MetaChainBlocks())
	assert.True(t, shardHeaders == tdp.ShardHeaders())
	assert.True(t, miniBlockheaders == tdp.MiniBlockHashes())
	assert.True(t, metaBlockNonce == tdp.MetaBlockNonces())
}
