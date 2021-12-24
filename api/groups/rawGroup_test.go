package groups_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
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
		GetRawBlockByNonceCalled: func(_ uint64, _ bool) (*block.MetaBlock, error) {
			return &block.MetaBlock{}, nil
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
		GetRawBlockByNonceCalled: func(_ uint64, _ bool) (*block.MetaBlock, error) {
			return &block.MetaBlock{}, nil
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
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedBlock := block.MetaBlock{
		Nonce: 15,
		Round: 17,
	}

	facade := mock.FacadeStub{
		GetRawBlockByNonceCalled: func(_ uint64, _ bool) (*block.MetaBlock, error) {
			return &expectedBlock, nil
		},
	}

	blockGroup, err := groups.NewRawBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/raw/block/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedBlock, response.Data.Block)
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
