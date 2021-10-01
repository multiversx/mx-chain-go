package groups_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlockGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewBlockGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewBlockGroup(&mock.Facade{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

type blockResponseData struct {
	Block api.Block `json:"block"`
}

type blockResponse struct {
	Data  blockResponseData `json:"data"`
	Error string            `json:"error"`
	Code  string            `json:"code"`
}

func TestGetBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*api.Block, error) {
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*api.Block, error) {
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockNonce.Error()))
}

func TestGetBlockByNonce_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*api.Block, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-nonce/37", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedBlock := api.Block{
		Nonce: 37,
		Round: 39,
	}
	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*api.Block, error) {
			return &expectedBlock, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-nonce/37", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedBlock, response.Data.Block)
}

// ---- by hash

func TestGetBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*api.Block, error) {
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetBlockByHash_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.Facade{
		GetBlockByHashCalled: func(_ string, _ bool) (*api.Block, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-hash/hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedBlock := api.Block{
		Nonce: 37,
		Round: 39,
	}
	facade := mock.Facade{
		GetBlockByHashCalled: func(_ string, _ bool) (*api.Block, error) {
			return &expectedBlock, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-hash/hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedBlock, response.Data.Block)
}

func getBlockRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"block": {
				Routes: []config.RouteConfig{
					{Name: "/by-nonce/:nonce", Open: true},
					{Name: "/by-hash/:hash", Open: true},
				},
			},
		},
	}
}
