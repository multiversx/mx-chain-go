package block_test

import (
	"encoding/hex"
	"encoding/json"
	errs "errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/errors"
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

type blockResponse struct {
	Nonce         uint64   `json:"nonce"`
	ShardID       uint32   `json:"shardId"`
	Hash          string   `json:"hash"`
	Proposer      string   `json:"proposer"`
	Validators    []string `json:"validators"`
	PubKeyBitmap  string   `json:"pubKeyBitmap"`
	Size          int64    `json:"size"`
	Timestamp     uint64   `json:"timestamp"`
	TxCount       uint32   `json:"txCount"`
	StateRootHash string   `json:"stateRootHash"`
	PrevHash      string   `json:"prevHash"`
}

type recentBlocksResponse struct {
	Blocks      []blockResponse `json:"blocks"`
	ShardHeader blockResponse   `json:"block"`
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

func startNodeServer(handler node.Handler) *gin.Engine {
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
	block.RoutesForBlockLists(blocksRoutes)
	return ws
}

//------- RecentBlocks

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
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestRecentBlocks_ReturnsCorrectly(t *testing.T) {
	t.Parallel()
	facade := mock.Facade{
		RecentNotarizedBlocksHandler: func(maxShardHeadersNum int) (blocks []*external.BlockHeader, e error) {
			return make([]*external.BlockHeader, 0), nil
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
	assert.Equal(t, 0, len(rb.Blocks))
}

//------- Block

func TestBlock_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing elrondFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", "test"), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestBlock_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", "test"), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := errorResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, errors.ErrInvalidAppContext.Error())
}

func TestBlock_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	testBlockHashHex := []byte("aaee")
	facade := mock.Facade{
		RetrieveShardBlockHandler: func(blockHash []byte) (info *external.ShardBlockInfo, e error) {
			blockHashConverted, _ := hex.DecodeString(string(testBlockHashHex))
			assert.Equal(t, blockHashConverted, blockHash)
			return &external.ShardBlockInfo{
				BlockHeader: external.BlockHeader{
					Nonce: 1,
				},
			}, nil
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", testBlockHashHex), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.NotNil(t, rb.ShardHeader)
	assert.Equal(t, uint64(1), rb.ShardHeader.Nonce)
}

func TestBlock_KeyNotFoundShouldReturnPageNotFound(t *testing.T) {
	t.Parallel()

	testBlockHashHex := []byte("aaee")
	facade := mock.Facade{
		RetrieveShardBlockHandler: func(blockHash []byte) (info *external.ShardBlockInfo, e error) {
			return &external.ShardBlockInfo{}, errs.New("not found")
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", testBlockHashHex), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusNotFound)
}

func TestBlock_KeyIsNotHexShouldReturnServerError(t *testing.T) {
	t.Parallel()

	testBlockHashHex := []byte("aae_")
	facade := mock.Facade{
		RetrieveShardBlockHandler: func(blockHash []byte) (info *external.ShardBlockInfo, e error) {
			return &external.ShardBlockInfo{}, errs.New("not found")
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", testBlockHashHex), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}
