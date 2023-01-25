package groups_test

import (
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
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
		hg, err := groups.NewBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

type alteredAccountsForBlockResponse struct {
	Data struct {
		Accounts []*outport.AlteredAccount `json:"accounts"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
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

	facade := mock.FacadeStub{
		GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
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

	facade := mock.FacadeStub{
		GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
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
	facade := mock.FacadeStub{
		GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
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
	facade := mock.FacadeStub{
		GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
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

	facade := mock.FacadeStub{
		GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
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
	facade := mock.FacadeStub{
		GetBlockByHashCalled: func(_ string, _ api.BlockQueryOptions) (*api.Block, error) {
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
	facade := mock.FacadeStub{
		GetBlockByHashCalled: func(_ string, _ api.BlockQueryOptions) (*api.Block, error) {
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
					{Name: "/by-round/:round", Open: true},
					{Name: "/altered-accounts/by-nonce/:nonce", Open: true},
					{Name: "/altered-accounts/by-hash/:hash", Open: true},
				},
			},
		},
	}
}

// ---- by round

func TestGetBlockByRound_WrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-round/2", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetBlockByRound_EmptyRoundUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-round", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetBlockByRound_InvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-round/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockRound.Error()))
}

func TestGetBlockByRound_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-round/37", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedBlock := api.Block{
		Nonce: 37,
		Round: 39,
	}
	facade := mock.FacadeStub{
		GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return &expectedBlock, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/block/by-round/37", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedBlock, response.Data.Block)
}

func TestGetBlockByRound_WithBadBlockQueryOptionsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	response, code := httpGetBlock(ws, "/block/by-round/37?withTxs=bad")
	require.Equal(t, http.StatusBadRequest, code)
	require.Contains(t, response.Error, apiErrors.ErrBadUrlParams.Error())

	response, code = httpGetBlock(ws, "/block/by-round/37?withLogs=bad")
	require.Equal(t, http.StatusBadRequest, code)
	require.Contains(t, response.Error, apiErrors.ErrBadUrlParams.Error())
}

func TestGetBlockByRound_WithBlockQueryOptionsShouldWork(t *testing.T) {
	t.Parallel()

	var calledWithRound uint64
	var calledWithOptions api.BlockQueryOptions

	facade := mock.FacadeStub{
		GetBlockByRoundCalled: func(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
			calledWithRound = round
			calledWithOptions = options
			return &api.Block{}, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	response, code := httpGetBlock(ws, "/block/by-round/37?withTxs=true")
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, response)
	require.Equal(t, uint64(37), calledWithRound)
	require.Equal(t, api.BlockQueryOptions{WithTransactions: true}, calledWithOptions)

	response, code = httpGetBlock(ws, "/block/by-round/38?withTxs=true&withLogs=true")
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, response)
	require.Equal(t, uint64(38), calledWithRound)
	require.Equal(t, api.BlockQueryOptions{WithTransactions: true, WithLogs: true}, calledWithOptions)
}

func TestGetAlteredAccountsByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedResponse := []*outport.AlteredAccount{
		{
			Address: "alice",
			Balance: "100000",
		},
	}

	facade := mock.FacadeStub{
		GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*outport.AlteredAccount, error) {
			require.Equal(t, api.BlockFetchTypeByNonce, options.RequestType)
			require.Equal(t, uint64(37), options.Nonce)

			return expectedResponse, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	response, code := httpGetAlteredAccountsForBlockBlock(ws, "/block/altered-accounts/by-nonce/37")
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, expectedResponse, response.Data.Accounts)
	require.Empty(t, response.Error)
	require.Equal(t, string(shared.ReturnCodeSuccess), response.Code)
}

func TestGetAlteredAccountsByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedResponse := []*outport.AlteredAccount{
		{
			Address: "alice",
			Balance: "100000",
		},
	}
	facade := mock.FacadeStub{
		GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*outport.AlteredAccount, error) {
			require.Equal(t, api.BlockFetchTypeByHash, options.RequestType)
			require.Equal(t, "aabb", hex.EncodeToString(options.Hash))

			return expectedResponse, nil
		},
	}

	blockGroup, err := groups.NewBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	response, code := httpGetAlteredAccountsForBlockBlock(ws, "/block/altered-accounts/by-hash/aabb")
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, expectedResponse, response.Data.Accounts)
	require.Empty(t, response.Error)
	require.Equal(t, string(shared.ReturnCodeSuccess), response.Code)
}

func httpGetBlock(ws *gin.Engine, url string) (blockResponse, int) {
	httpRequest, _ := http.NewRequest("GET", url, nil)
	httpResponse := httptest.NewRecorder()
	ws.ServeHTTP(httpResponse, httpRequest)

	blockResponse := blockResponse{}
	loadResponse(httpResponse.Body, &blockResponse)
	return blockResponse, httpResponse.Code
}

func httpGetAlteredAccountsForBlockBlock(ws *gin.Engine, url string) (alteredAccountsForBlockResponse, int) {
	httpRequest, _ := http.NewRequest("GET", url, nil)
	httpResponse := httptest.NewRecorder()
	ws.ServeHTTP(httpResponse, httpRequest)

	response := alteredAccountsForBlockResponse{}
	loadResponse(httpResponse.Body, &response)
	return response, httpResponse.Code
}
