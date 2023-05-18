package groups_test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
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
		Accounts []*alteredAccount.AlteredAccount `json:"accounts"`
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

func TestBlockGroup_getBlockByNonce(t *testing.T) {
	t.Parallel()

	t.Run("empty nonce should error", func(t *testing.T) {
		t.Parallel()

		testBlockGroup(t, &mock.FacadeStub{}, "/block/by-nonce", nil, http.StatusNotFound, "")
	})
	t.Run("invalid nonce should error",
		testBlockGroupErrorScenario("/block/by-nonce/invalid", nil, formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockNonce)))
	t.Run("invalid query options should error",
		testBlockGroupErrorScenario("/block/by-nonce/10?withTxs=not-bool", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetBlockByNonceCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
				return nil, expectedErr
			},
		}

		testBlockGroup(
			t,
			facade,
			"/block/by-nonce/10",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedNonce := uint64(37)
		expectedOptions := api.BlockQueryOptions{WithTransactions: true}
		expectedBlock := api.Block{
			Nonce: 37,
			Round: 39,
		}
		facade := &mock.FacadeStub{
			GetBlockByNonceCalled: func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
				require.Equal(t, providedNonce, nonce)
				require.Equal(t, expectedOptions, options)
				return &expectedBlock, nil
			},
		}

		response := &blockResponse{}
		loadBlockGroupResponse(
			t,
			facade,
			fmt.Sprintf("/block/by-nonce/%d?withTxs=true", providedNonce),
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedBlock, response.Data.Block)
	})
}

func TestBlockGroup_getBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testBlockGroup(t, &mock.FacadeStub{}, "/block/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("invalid query options should error",
		testBlockGroupErrorScenario("/block/by-hash/hash?withLogs=not-bool", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetBlockByHashCalled: func(_ string, _ api.BlockQueryOptions) (*api.Block, error) {
				return nil, expectedErr
			},
		}

		testBlockGroup(
			t,
			facade,
			"/block/by-hash/hash",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHash := "hash"
		expectedOptions := api.BlockQueryOptions{WithTransactions: true}
		expectedBlock := api.Block{
			Nonce: 37,
			Round: 39,
		}
		facade := &mock.FacadeStub{
			GetBlockByHashCalled: func(hash string, options api.BlockQueryOptions) (*api.Block, error) {
				require.Equal(t, providedHash, hash)
				require.Equal(t, expectedOptions, options)
				return &expectedBlock, nil
			},
		}

		response := &blockResponse{}
		loadBlockGroupResponse(
			t,
			facade,
			fmt.Sprintf("/block/by-hash/%s?withTxs=true", providedHash),
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedBlock, response.Data.Block)
	})
}

func TestBlockGroup_getBlockByRound(t *testing.T) {
	t.Parallel()

	t.Run("empty round should error", func(t *testing.T) {
		t.Parallel()

		testBlockGroup(t, &mock.FacadeStub{}, "/block/by-round", nil, http.StatusNotFound, "")
	})
	t.Run("invalid round should error",
		testBlockGroupErrorScenario("/block/by-round/invalid", nil, formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockRound)))
	t.Run("invalid query options should error",
		testBlockGroupErrorScenario("/block/by-round/123?withTxs=not-bool", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrBadUrlParams)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
				return nil, expectedErr
			},
		}

		testBlockGroup(
			t,
			facade,
			"/block/by-round/123",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedRound := uint64(37)
		expectedOptions := api.BlockQueryOptions{WithTransactions: true}
		expectedBlock := api.Block{
			Nonce: 37,
			Round: 39,
		}
		facade := &mock.FacadeStub{
			GetBlockByRoundCalled: func(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
				require.Equal(t, providedRound, round)
				require.Equal(t, expectedOptions, options)
				return &expectedBlock, nil
			},
		}

		response := &blockResponse{}
		loadBlockGroupResponse(
			t,
			facade,
			fmt.Sprintf("/block/by-round/%d?withTxs=true", providedRound),
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedBlock, response.Data.Block)
	})
}

func TestBlockGroup_getAlteredAccountsByNonce(t *testing.T) {
	t.Parallel()

	t.Run("empty nonce should error", func(t *testing.T) {
		t.Parallel()

		testBlockGroup(t, &mock.FacadeStub{}, "/block/altered-accounts/by-nonce", nil, http.StatusNotFound, "")
	})
	t.Run("invalid nonce should error",
		testBlockGroupErrorScenario("/block/altered-accounts/by-nonce/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetAlteredAccountsForBlock, apiErrors.ErrInvalidBlockRound)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
				return nil, expectedErr
			},
		}

		testBlockGroup(
			t,
			facade,
			"/block/altered-accounts/by-nonce/123",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetAlteredAccountsForBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedNonce := uint64(37)
		expectedOptions := api.GetAlteredAccountsForBlockOptions{
			GetBlockParameters: api.GetBlockParameters{
				RequestType: api.BlockFetchTypeByNonce,
				Nonce:       providedNonce,
			},
		}
		expectedResponse := []*alteredAccount.AlteredAccount{
			{
				Address: "alice",
				Balance: "100000",
			},
		}

		facade := &mock.FacadeStub{
			GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
				require.Equal(t, expectedOptions, options)
				return expectedResponse, nil
			},
		}

		response := &alteredAccountsForBlockResponse{}
		loadBlockGroupResponse(
			t,
			facade,
			fmt.Sprintf("/block/altered-accounts/by-nonce/%d", providedNonce),
			"GET",
			nil,
			response,
		)
		require.Equal(t, expectedResponse, response.Data.Accounts)
		require.Empty(t, response.Error)
		require.Equal(t, string(shared.ReturnCodeSuccess), response.Code)
	})
}

func TestBlockGroup_getAlteredAccountsByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testBlockGroup(t, &mock.FacadeStub{}, "/block/altered-accounts/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("invalid hash should error",
		testBlockGroupErrorScenario("/block/altered-accounts/by-hash/hash", nil,
			apiErrors.ErrGetAlteredAccountsForBlock.Error()))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		providedHash := hex.EncodeToString([]byte("hash"))
		facade := &mock.FacadeStub{
			GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
				return nil, expectedErr
			},
		}

		testBlockGroup(
			t,
			facade,
			"/block/altered-accounts/by-hash/"+providedHash,
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetAlteredAccountsForBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHash := hex.EncodeToString([]byte("hash"))
		expectedOptions := api.GetAlteredAccountsForBlockOptions{
			GetBlockParameters: api.GetBlockParameters{
				RequestType: api.BlockFetchTypeByHash,
				Hash:        []byte("hash"),
			},
		}
		expectedResponse := []*alteredAccount.AlteredAccount{
			{
				Address: "alice",
				Balance: "100000",
			},
		}

		facade := &mock.FacadeStub{
			GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
				require.Equal(t, providedHash, hex.EncodeToString(options.Hash))
				require.Equal(t, expectedOptions, options)
				return expectedResponse, nil
			},
		}

		response := &alteredAccountsForBlockResponse{}
		loadBlockGroupResponse(
			t,
			facade,
			fmt.Sprintf("/block/altered-accounts/by-hash/%s", providedHash),
			"GET",
			nil,
			response,
		)
		require.Equal(t, expectedResponse, response.Data.Accounts)
		require.Empty(t, response.Error)
		require.Equal(t, string(shared.ReturnCodeSuccess), response.Code)
	})
}

func TestBlockGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	blockGroup, _ := groups.NewBlockGroup(nil)
	require.True(t, blockGroup.IsInterfaceNil())

	blockGroup, _ = groups.NewBlockGroup(&mock.FacadeStub{})
	require.False(t, blockGroup.IsInterfaceNil())
}

func TestBlockGroup_UpdateFacadeStub(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		blockGroup, err := groups.NewBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = blockGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		blockGroup, err := groups.NewBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = blockGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedBlock := api.Block{
			Nonce: 37,
			Round: 39,
		}
		facade := mock.FacadeStub{
			GetBlockByNonceCalled: func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
				return &expectedBlock, nil
			},
		}

		blockGroup, err := groups.NewBlockGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

		req, _ := http.NewRequest("GET", "/block/by-nonce/10", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := blockResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, expectedBlock, response.Data.Block)

		newFacade := mock.FacadeStub{
			GetBlockByNonceCalled: func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
				return nil, expectedErr
			},
		}
		err = blockGroup.UpdateFacade(&newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/block/by-nonce/10", nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response = blockResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
}

func loadBlockGroupResponse(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	destination interface{},
) {
	blockGroup, err := groups.NewBlockGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	loadResponse(resp.Body, destination)
}

func testBlockGroupErrorScenario(url string, body io.Reader, expectedErr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		testBlockGroup(
			t,
			&mock.FacadeStub{},
			url,
			body,
			http.StatusBadRequest,
			expectedErr,
		)
	}
}

func testBlockGroup(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	body io.Reader,
	expectedRespCode int,
	expectedRespError string,
) {
	blockGroup, err := groups.NewBlockGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "block", getBlockRoutesConfig())

	req, _ := http.NewRequest("GET", url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, expectedRespCode, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedRespError))
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
