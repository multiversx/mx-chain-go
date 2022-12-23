package groups_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rawBlockResponseData struct {
	Block []byte `json:"block"`
}

type rawBlockResponse struct {
	Data  rawBlockResponseData `json:"data"`
	Error string               `json:"error"`
	Code  string               `json:"code"`
}

type internalMetaBlockResponseData struct {
	Block block.MetaBlock `json:"block"`
}

type internalMetaBlockResponse struct {
	Data  internalMetaBlockResponseData `json:"data"`
	Error string                        `json:"error"`
	Code  string                        `json:"code"`
}

type internalShardBlockResponseData struct {
	Block block.Header `json:"block"`
}

type internalShardBlockResponse struct {
	Data  internalShardBlockResponseData `json:"data"`
	Error string                         `json:"error"`
	Code  string                         `json:"code"`
}

type rawMiniBlockResponseData struct {
	Block []byte `json:"miniblock"`
}

type rawMiniBlockResponse struct {
	Data  rawMiniBlockResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}

type internalMiniBlockResponseData struct {
	Block block.MiniBlock `json:"miniblock"`
}

type internalMiniBlockResponse struct {
	Data  internalMiniBlockResponseData `json:"data"`
	Error string                        `json:"error"`
	Code  string                        `json:"code"`
}

type internalValidatorsInfoResponse struct {
	Data struct {
		ValidatorsInfo []*state.ShardValidatorInfo `json:"validators"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func TestNewInternalBlockGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewInternalBlockGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewInternalBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

// ---- RAW

// ---- MetaBlock - by nonce

func TestGetRawMetaBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMetaBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockNonce.Error()))
}

func TestGetRawMetaBlockByNonce_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetRawMetaBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MetaBlock - by round

func TestGetRawMetaBlockByRound_EmptyRoundUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-round", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMetaBlockByRound_InvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-round/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockRound.Error()))
}

func TestGetRawMetaBlockByRound_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetRawMetaBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MetaBlock - by hash

func TestGetRawMetaBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMetaBlockByHash_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-hash/dummyhash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetRawMetaBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- StartOfEpoch MetaBlock - raw

func TestGetRawStartOfEpochMetaBlock_NoEpochUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/startofepoch/metablock/by-epoch/a", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawStartOfEpochMetaBlock_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/startofepoch/metablock/by-epoch/1", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetRawStartOfEpochMetaBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/startofepoch/metablock/by-epoch/1", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ----------------- Shard Block ---------------

func TestGetRawShardBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawShardBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawShardBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- ShardBlock - by round

func TestGetRawShardBlockByRound_EmptyRoundUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-round", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawShardBlockByRound_InvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-round/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockRound.Error()))
}

func TestGetRawShardBlockByRound_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetRawShardBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- ShardBlock - by hash

func TestGetRawShardBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawShardBlockByHash_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-hash/dummyhash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetRawShardBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MiniBlock

func TestGetRawMiniBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/miniblock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMiniBlockByHash_NoEpochUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/miniblock/by-hash/aaaa/epoch", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMiniBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/miniblock/by-hash/aaaa/epoch/1", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- JSON

// ---- MetaBlock - by nonce

func TestGetInternalMetaBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalMetaBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockNonce.Error()))
}

func TestGetInternalMetaBlockByNonce_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetInternalMetaBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.MetaBlock{
		Nonce: 15,
		Epoch: 15,
	}

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MetaBlock - by round

func TestGetInternalMetaBlockByRound_EmptyRoundUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-round", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalMetaBlockByRound_InvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-round/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockRound.Error()))
}

func TestGetInternalMetaBlockByRound_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetInternalMetaBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.MetaBlock{
		Nonce: 15,
		Epoch: 15,
	}

	facade := mock.FacadeStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MetaBlock - by hash

func TestGetInternalMetaBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalMetaBlockByHash_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-hash/dummyhash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetInternalMetaBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.MetaBlock{
		Nonce: 15,
		Epoch: 15,
	}

	facade := mock.FacadeStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/metablock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMetaBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- StartOfEpoch MetaBlock - json

func TestGetInternalStartOfEpochMetaBlock_NoEpochUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/startofepoch/metablock/by-epoch", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalStartOfEpochMetaBlock_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/startofepoch/metablock/by-epoch/1", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetInternalStartOfEpochMetaBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/startofepoch/metablock/by-epoch/1", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ----------------- Shard Block ---------------

func TestGetInternalShardBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalShardBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetInternalShardBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.Header{
		Nonce: 15,
		Round: 15,
	}

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- ShardBlock - by round

func TestGetInternalShardBlockByRound_EmptyRoundUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-round", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalShardBlockByRound_InvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-round/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrInvalidBlockRound.Error()))
}

func TestGetInternalShardBlockByRound_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetInternalShardBlockByRound_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.Header{
		Nonce: 15,
		Round: 15,
	}

	facade := mock.FacadeStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-round/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- ShardBlock - by hash

func TestGetInternalShardBlockByHash_NoHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalShardBlockByHash_FacadeErrorShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("local err")
	facade := mock.FacadeStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return nil, expectedErr
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-hash/dummyhash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
}

func TestGetInternalShardBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.Header{
		Nonce: 15,
		Round: 15,
	}

	facade := mock.FacadeStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/shardblock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalShardBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MiniBlock

func TestGetInternalMiniBlockByHash_EmptyHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/miniblock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalMiniBlockByHash_NoEpochUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/miniblock/by-hash/aaaa/epoch", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetInternalMiniBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := block.MiniBlock{}

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/json/miniblock/by-hash/dummyhash/epoch/1", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := internalMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

func TestGetInternalStartOfEpochValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("no epoch param should fail", func(t *testing.T) {
		t.Parallel()

		facade := mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return make([]*state.ShardValidatorInfo, 0), nil
			},
		}

		blockGroup, err := groups.NewInternalBlockGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

		req, _ := http.NewRequest("GET", "/internal/json/startofepoch/validators/by-epoch/aaa", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := internalValidatorsInfoResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.True(t, strings.Contains(response.Error, apiErrors.ErrGetValidatorsInfo.Error()))
	})

	t.Run("facade error should fail", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("facade error")
		facade := mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return nil, expectedErr
			},
		}

		blockGroup, err := groups.NewInternalBlockGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

		req, _ := http.NewRequest("GET", "/internal/json/startofepoch/validators/by-epoch/1", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := internalValidatorsInfoResponse{}
		loadResponse(resp.Body, &response)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedOutput := []*state.ShardValidatorInfo{
			{
				PublicKey:  []byte("pubkey1"),
				ShardId:    0,
				Index:      1,
				TempRating: 500,
			},
		}

		facade := mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return expectedOutput, nil
			},
		}

		blockGroup, err := groups.NewInternalBlockGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

		req, _ := http.NewRequest("GET", "/internal/json/startofepoch/validators/by-epoch/1", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := internalValidatorsInfoResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)

		assert.Equal(t, expectedOutput, response.Data.ValidatorsInfo)
	})

}

func getInternalBlockRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"internal": {
				Routes: []config.RouteConfig{
					{Name: "/raw/metablock/by-nonce/:nonce", Open: true},
					{Name: "/raw/metablock/by-hash/:hash", Open: true},
					{Name: "/raw/metablock/by-round/:round", Open: true},
					{Name: "/raw/startofepoch/metablock/by-epoch/:epoch", Open: true},
					{Name: "/raw/shardblock/by-nonce/:nonce", Open: true},
					{Name: "/raw/shardblock/by-hash/:hash", Open: true},
					{Name: "/raw/shardblock/by-round/:round", Open: true},
					{Name: "/raw/miniblock/by-hash/:hash/epoch/:epoch", Open: true},
					{Name: "/json/metablock/by-nonce/:nonce", Open: true},
					{Name: "/json/metablock/by-hash/:hash", Open: true},
					{Name: "/json/metablock/by-round/:round", Open: true},
					{Name: "/json/startofepoch/metablock/by-epoch/:epoch", Open: true},
					{Name: "/json/shardblock/by-nonce/:nonce", Open: true},
					{Name: "/json/shardblock/by-hash/:hash", Open: true},
					{Name: "/json/shardblock/by-round/:round", Open: true},
					{Name: "/json/miniblock/by-hash/:hash/epoch/:epoch", Open: true},
					{Name: "/json/startofepoch/validators/by-epoch/:epoch", Open: true},
				},
			},
		},
	}
}
