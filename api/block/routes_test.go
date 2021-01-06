package block_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/block"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type blockResponseData struct {
	Block block.APIBlock `json:"block"`
}

type blockResponse struct {
	Data  blockResponseData `json:"data"`
	Error string            `json:"error"`
	Code  string            `json:"code"`
}

func TestGetBlockByNonce_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/block/by-nonce/5", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetBlockByNonce_WrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	req, _ := http.NewRequest("GET", "/block/by-nonce/2", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidAppContext.Error()))
}

func TestGetBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*block.APIBlock, error) {
			return &block.APIBlock{}, nil
		},
	}

	ws := startNodeServer(&facade)

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
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*block.APIBlock, error) {
			return &block.APIBlock{}, nil
		},
	}

	ws := startNodeServer(&facade)

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
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*block.APIBlock, error) {
			return nil, expectedErr
		},
	}

	ws := startNodeServer(&facade)

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

	expectedBlock := block.APIBlock{
		Nonce: 37,
		Round: 39,
	}
	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*block.APIBlock, error) {
			return &expectedBlock, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/block/by-nonce/37", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedBlock, response.Data.Block)
}

// ---- by hash

func TestGetBlockByHash_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/block/by-hash/5", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetBlockByHash_WrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()

	req, _ := http.NewRequest("GET", "/block/by-hash/2", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidAppContext.Error()))
}

func TestGetBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetBlockByNonceCalled: func(_ uint64, _ bool) (*block.APIBlock, error) {
			return &block.APIBlock{}, nil
		},
	}

	ws := startNodeServer(&facade)

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
		GetBlockByHashCalled: func(_ string, _ bool) (*block.APIBlock, error) {
			return nil, expectedErr
		},
	}

	ws := startNodeServer(&facade)

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

	expectedBlock := block.APIBlock{
		Nonce: 37,
		Round: 39,
	}
	facade := mock.Facade{
		GetBlockByHashCalled: func(_ string, _ bool) (*block.APIBlock, error) {
			return &expectedBlock, nil
		},
	}

	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/block/by-hash/hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedBlock, response.Data.Block)
}

func startNodeServer(handler block.BlockService) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	blockRoutes := ws.Group("/block")
	if handler != nil {
		blockRoutes.Use(middleware.WithFacade(handler))
	}
	blockRoute, _ := wrapper.NewRouterWrapper("block", blockRoutes, getRoutesConfig())
	block.Routes(blockRoute)
	return ws
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"block": {
				[]config.RouteConfig{
					{Name: "/by-nonce/:nonce", Open: true},
					{Name: "/by-hash/:hash", Open: true},
				},
			},
		},
	}
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	logError(err)
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("facade", mock.WrongFacade{})
	})
	ginBlockRoute := ws.Group("/block")
	blockRoute, _ := wrapper.NewRouterWrapper("block", ginBlockRoute, getRoutesConfig())
	block.Routes(blockRoute)
	return ws
}
