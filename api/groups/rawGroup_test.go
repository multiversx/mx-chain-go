package groups_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	marshalizerFactory "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
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

func TestGetRawBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/block/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/block/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawBlockByNonce_ShouldWork(t *testing.T) {
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

	req, _ := http.NewRequest("GET", "/raw/block/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, bytes.Repeat([]byte("1"), 10), response.Data.Block)
}

func TestGetRawBlockByNonceMetaBlockCheck_ShouldWork(t *testing.T) {
	t.Parallel()

	metaBlock := block.MetaBlock{
		Nonce: 15,
		Round: 17,
	}

	marshalizer, err := marshalizerFactory.NewMarshalizer("gogo protobuf")
	require.NoError(t, err)

	// expectedBlock, err := marshalizer.Marshal(metaBlock)
	// require.NoError(t, err)

	facade := mock.FacadeStub{
		GetRawMetaBlockByNonceCalled: func(_ uint64, _ bool) ([]byte, error) {
			//return bytes.Repeat([]byte("1"), 10), nil
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/block/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	blockHeader := &block.MetaBlock{}
	err = marshalizer.Unmarshal(blockHeader, response.Data.Block)
	require.NoError(t, err)

	assert.Equal(t, metaBlock, blockHeader)

	assert.Equal(t, bytes.Repeat([]byte("1"), 10), response.Data.Block)
}

func getRawBlockRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"raw": {
				Routes: []config.RouteConfig{
					{Name: "/block/by-nonce/:nonce", Open: true},
					{Name: "/block/by-hash/:hash", Open: true},
					{Name: "/block/by-round/:round", Open: true},
				},
			},
		},
	}
}
