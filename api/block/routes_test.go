package block_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/block"
	apiErrors "github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/external"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type errorResponse struct {
	Error string `json:"error"`
}

type recentBlocksResponse struct {
	errorResponse
	Blocks []external.RecentBlock `json:"blocks"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler node.FacadeHandler) *gin.Engine {
	server := startNodeServerWithFacade(handler)
	return server
}

func startNodeServerWrongFacade() *gin.Engine {
	return startNodeServerWithFacade(mock.WrongFacade{})
}

func startNodeServerWithFacade(facade interface{}) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	if facade != nil {
		ws.Use(func(c *gin.Context) {
			c.Set("elrondFacade", facade)
		})
	}

	blockRoutes := ws.Group("/block")
	block.Routes(blockRoutes)
	blocksRoutes := ws.Group("/blocks")
	block.RoutesForBlocksLists(blocksRoutes)
	return ws
}

func TestRecentBlocks_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestRecentBlocks_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := errorResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestRecentBlocks_ReturnsCorrectly(t *testing.T) {
	t.Parallel()
	recentBlocks := []external.RecentBlock{
		{Nonce: 0, Hash: make([]byte, 0), PrevHash: make([]byte, 0), StateRootHash: make([]byte, 0)},
		{Nonce: 0, Hash: make([]byte, 0), PrevHash: make([]byte, 0), StateRootHash: make([]byte, 0)},
	}
	facade := mock.Facade{
		RecentNotarizedBlocksHandler: func(maxShardHeadersNum int) (blocks []external.RecentBlock, e error) {
			return recentBlocks, nil
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.NotNil(t, rb.Blocks)
	assert.Equal(t, recentBlocks, rb.Blocks)
}

func TestRecentBlocks_ReturnsErrorWhenRecentBlocksErrors(t *testing.T) {
	t.Parallel()
	errMessage := "recent blocks error"
	facade := mock.Facade{
		RecentNotarizedBlocksHandler: func(maxShardHeadersNum int) (blocks []external.RecentBlock, e error) {
			return nil, errors.New(errMessage)
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Nil(t, rb.Blocks)
	assert.NotNil(t, rb.Error)
	assert.Equal(t, errMessage, rb.Error)
}
