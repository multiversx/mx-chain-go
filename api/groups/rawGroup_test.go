package groups_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRawBlockGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewRawBlockGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewRawBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestGetRawMetaBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/metablock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMetaBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/metablock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawMetaBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	// metaBlock := block.MetaBlock{
	// 	Nonce: 15,
	// 	Round: 17,
	// }

	// marshalizer := mock.MarshalizerStub{}

	// expectedBlock, err := marshalizer.Marshal(metaBlock)
	// require.NoError(t, err)

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return bytes.Repeat([]byte("1"), 10), nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, bytes.Repeat([]byte("1"), 10), response.Data.Block)
}

func TestGetRawMetaBlockByNonceMetaBlockCheck_ShouldWork(t *testing.T) {
	t.Parallel()

	// metaBlock := block.MetaBlock{
	// 	Nonce: 15,
	// 	Round: 17,
	// }

	// 	marshalizer, err := marshalizerFactory.NewMarshalizer("gogo protobuf")
	// 	require.NoError(t, err)

	// expectedBlock, err := marshalizer.Marshal(metaBlock)
	// require.NoError(t, err)

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return bytes.Repeat([]byte("1"), 10), nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	// blockHeader := &block.MetaBlock{}
	// err = marshalizer.Unmarshal(blockHeader, response.Data.Block)
	// require.NoError(t, err)

	// assert.Equal(t, metaBlock, blockHeader)

	assert.Equal(t, bytes.Repeat([]byte("1"), 10), response.Data.Block)
}

// ----------------- Shard Block ---------------

func TestGetRawShardBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetRawShardBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/shardblock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawShardBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetRawShardBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/shardblock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawShardBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	// metaBlock := block.ShardBlock{
	// 	Nonce: 15,
	// 	Round: 17,
	// }

	// marshalizer := mock.MarshalizerStub{}

	// expectedBlock, err := marshalizer.Marshal(metaBlock)
	// require.NoError(t, err)

	facade := mock.FacadeStub{
		GetRawShardBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return bytes.Repeat([]byte("1"), 10), nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/shardblock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, bytes.Repeat([]byte("1"), 10), response.Data.Block)
}

func TestGetRawShardBlockByNonceMetaBlockCheck_ShouldWork(t *testing.T) {
	t.Parallel()

	// metaBlock := block.ShardBlock{
	// 	Nonce: 15,
	// 	Round: 17,
	// }

	// 	marshalizer, err := marshalizerFactory.NewMarshalizer("gogo protobuf")
	// 	require.NoError(t, err)

	// expectedBlock, err := marshalizer.Marshal(metaBlock)
	// require.NoError(t, err)

	facade := mock.FacadeStub{
		GetRawShardBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return bytes.Repeat([]byte("1"), 10), nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/shardblock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	// blockHeader := &block.ShardBlock{}
	// err = marshalizer.Unmarshal(blockHeader, response.Data.Block)
	// require.NoError(t, err)

	// assert.Equal(t, metaBlock, blockHeader)

	assert.Equal(t, bytes.Repeat([]byte("1"), 10), response.Data.Block)
}

func getRawBlockRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"raw": {
				Routes: []config.RouteConfig{
					{Name: "/metablock/by-nonce/:nonce", Open: true},
					{Name: "/metablock/by-hash/:hash", Open: true},
					{Name: "/metablock/by-round/:round", Open: true},
					{Name: "/shardblock/by-nonce/:nonce", Open: true},
					{Name: "/shardblock/by-hash/:hash", Open: true},
					{Name: "/shardblock/by-round/:round", Open: true},
				},
			},
		},
	}
}
