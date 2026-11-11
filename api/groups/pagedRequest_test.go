package groups

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/middleware"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func createPagedContext(method string, query string, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/address/alice/keys?"+query, strings.NewReader(body))

	return c, w
}

func TestExtractPagedRequest(t *testing.T) {
	t.Parallel()

	epoch := uint32(3)
	cp := &checkpoint.Checkpoint{BlockNonce: 10, BlockHash: "aa", BlockRootHash: "bb", Epoch: &epoch, IteratorState: [][]byte{[]byte("state")}}
	encodedCp, err := checkpoint.Encode(cp)
	require.Nil(t, err)
	cpJSON, err := json.Marshal(cp)
	require.Nil(t, err)

	t.Run("no parameters", func(t *testing.T) {
		t.Parallel()

		for _, method := range []string{http.MethodGet, http.MethodPost} {
			c, _ := createPagedContext(method, "", "")
			request, err := ExtractPagedRequest(c)
			require.Nil(t, err, method)
			require.Equal(t, PagedRequest{}, request)
		}
	})
	t.Run("query parameters", func(t *testing.T) {
		t.Parallel()

		for _, method := range []string{http.MethodGet, http.MethodPost} {
			c, _ := createPagedContext(method, "numKeys=7&iteratorState="+encodedCp+"&blockRootHash=cc&hintEpoch=4", "")
			request, err := ExtractPagedRequest(c)
			require.Nil(t, err, method)
			require.Equal(t, uint(7), request.NumKeys)
			require.Equal(t, cp, request.IteratorState)
			require.Equal(t, []byte{0xcc}, request.Options.BlockRootHash)
			require.Equal(t, core.OptionalUint32{Value: 4, HasValue: true}, request.Options.HintEpoch)
		}
	})
	t.Run("body parameters", func(t *testing.T) {
		t.Parallel()

		body := `{"numKeys":5,"iteratorState":` + string(cpJSON) + `,"blockNonce":12,"onFinalBlock":true,"onStartOfEpoch":2,"hintEpoch":1,"blockHash":"dd","blockRootHash":"ee"}`
		c, _ := createPagedContext(http.MethodPost, "", body)
		request, err := ExtractPagedRequest(c)
		require.ErrorIs(t, err, apiErrors.ErrBadUrlParams) // options incompatible, parsed from the body
		require.Equal(t, PagedRequest{}, request)

		c, _ = createPagedContext(http.MethodPost, "", `{"numKeys":5,"iteratorState":`+string(cpJSON)+`,"blockRootHash":"ee","hintEpoch":1}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(5), request.NumKeys)
		require.Equal(t, cp, request.IteratorState)
		require.Equal(t, []byte{0xee}, request.Options.BlockRootHash)
		require.Equal(t, core.OptionalUint32{Value: 1, HasValue: true}, request.Options.HintEpoch)

		c, _ = createPagedContext(http.MethodPost, "", `{"blockNonce":12}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, core.OptionalUint64{Value: 12, HasValue: true}, request.Options.BlockNonce)

		c, _ = createPagedContext(http.MethodPost, "", `{"onFinalBlock":true,"onStartOfEpoch":2}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.True(t, request.Options.OnFinalBlock)
		require.Equal(t, core.OptionalUint32{Value: 2, HasValue: true}, request.Options.OnStartOfEpoch)

		c, _ = createPagedContext(http.MethodPost, "", `{"blockHash":"dd"}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, []byte{0xdd}, request.Options.BlockHash)
	})
	t.Run("body overrides query", func(t *testing.T) {
		t.Parallel()

		c, _ := createPagedContext(http.MethodPost, "numKeys=7&iteratorState="+encodedCp+"&blockNonce=1", `{"numKeys":5,"blockNonce":2,"iteratorState":{"blockNonce":99,"iteratorState":null}}`)
		request, err := ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(5), request.NumKeys)
		require.Equal(t, core.OptionalUint64{Value: 2, HasValue: true}, request.Options.BlockNonce)
		require.Equal(t, &checkpoint.Checkpoint{BlockNonce: 99}, request.IteratorState)

		c, _ = createPagedContext(http.MethodPost, "numKeys=7&iteratorState="+encodedCp, `{"iteratorState":null}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(7), request.NumKeys)
		require.Equal(t, cp, request.IteratorState)
	})
	t.Run("body zero values override, null and empty hashes do not", func(t *testing.T) {
		t.Parallel()

		c, _ := createPagedContext(http.MethodPost, "numKeys=7&onFinalBlock=true&blockHash=dd", `{"numKeys":0,"onFinalBlock":false,"blockHash":""}`)
		request, err := ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(0), request.NumKeys)
		require.False(t, request.Options.OnFinalBlock)
		require.Equal(t, []byte{0xdd}, request.Options.BlockHash)

		c, _ = createPagedContext(http.MethodPost, "numKeys=7&onFinalBlock=true", `{"numKeys":null,"onFinalBlock":null,"blockHash":""}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(7), request.NumKeys)
		require.True(t, request.Options.OnFinalBlock)
		require.Nil(t, request.Options.BlockHash)

		c, _ = createPagedContext(http.MethodPost, "blockHash=dd", `{"blockHash":null,"blockRootHash":""}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, []byte{0xdd}, request.Options.BlockHash)
		require.Nil(t, request.Options.BlockRootHash)

		c, _ = createPagedContext(http.MethodPost, "blockRootHash=ee&hintEpoch=3", `{"blockNonce":null,"onStartOfEpoch":null,"hintEpoch":null,"blockRootHash":null}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, []byte{0xee}, request.Options.BlockRootHash)
		require.Equal(t, core.OptionalUint32{Value: 3, HasValue: true}, request.Options.HintEpoch)
		require.False(t, request.Options.BlockNonce.HasValue)
		require.False(t, request.Options.OnStartOfEpoch.HasValue)

		c, _ = createPagedContext(http.MethodPost, "blockNonce=5", `{"blockNonce":0}`)
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, core.OptionalUint64{Value: 0, HasValue: true}, request.Options.BlockNonce)
	})
	t.Run("nil body and non json content type", func(t *testing.T) {
		t.Parallel()

		c, _ := createPagedContext(http.MethodPost, "numKeys=7", "")
		c.Request.Body = nil
		request, err := ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(7), request.NumKeys)

		c, _ = createPagedContext(http.MethodPost, "numKeys=7", `{"numKeys":5}`)
		c.Request.Header.Set("Content-Type", "text/plain")
		request, err = ExtractPagedRequest(c)
		require.Nil(t, err)
		require.Equal(t, uint(5), request.NumKeys)
	})
	t.Run("body is read only on POST", func(t *testing.T) {
		t.Parallel()

		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			c, _ := createPagedContext(method, "numKeys=7", `{"numKeys":5}`)
			request, err := ExtractPagedRequest(c)
			require.Nil(t, err, method)
			require.Equal(t, uint(7), request.NumKeys, method)
		}
	})
	t.Run("options are checked after the merge", func(t *testing.T) {
		t.Parallel()

		c, _ := createPagedContext(http.MethodPost, "blockNonce=1", `{"blockHash":"dd"}`)
		_, err := ExtractPagedRequest(c)
		require.ErrorIs(t, err, apiErrors.ErrBadUrlParams)
		require.Contains(t, err.Error(), "only one block coordinate")

		c, _ = createPagedContext(http.MethodGet, "blockNonce=1&blockHash=dd", "")
		_, err = ExtractPagedRequest(c)
		require.ErrorIs(t, err, apiErrors.ErrBadUrlParams)
	})
	t.Run("invalid query parameters", func(t *testing.T) {
		t.Parallel()

		for _, query := range []string{"numKeys=-1", "numKeys=abc", "numKeys=99999999999999999999", "iteratorState=%25%25", "iteratorState=bm90IGpzb24", "blockNonce=abc", "blockHash=zz"} {
			c, _ := createPagedContext(http.MethodGet, query, "")
			request, err := ExtractPagedRequest(c)
			require.ErrorIs(t, err, apiErrors.ErrBadUrlParams, query)
			require.Equal(t, PagedRequest{}, request)
		}
	})
	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()

		for _, body := range []string{"not json", `{"numKeys":`, `{"numKeys":"a"}`, `{"numKeys":-1}`, `{"numKeys":99999999999999999999}`, `{"blockNonce":-1}`, `{"iteratorState":{"iteratorState":["%%"]}}`, `{"blockHash":"zz"}`, `{"blockRootHash":"zz"}`} {
			c, _ := createPagedContext(http.MethodPost, "numKeys=7", body)
			request, err := ExtractPagedRequest(c)
			require.ErrorIs(t, err, apiErrors.ErrValidation, body)
			require.Equal(t, PagedRequest{}, request)
		}
	})
}

func TestRespondWithPage(t *testing.T) {
	t.Parallel()

	t.Run("without iterator state", func(t *testing.T) {
		t.Parallel()

		c, w := createPagedContext(http.MethodGet, "", "")
		RespondWithPage(c, gin.H{"pairs": map[string]string{"aa": "bb"}}, nil)
		require.Equal(t, http.StatusOK, w.Code)
		response := shared.GenericAPIResponse{}
		require.Nil(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Equal(t, shared.ReturnCodeSuccess, response.Code)
		require.Equal(t, map[string]interface{}{"pairs": map[string]interface{}{"aa": "bb"}}, response.Data)
	})
	t.Run("nil data with iterator state", func(t *testing.T) {
		t.Parallel()

		c, w := createPagedContext(http.MethodGet, "", "")
		RespondWithPage(c, nil, &checkpoint.Checkpoint{BlockNonce: 1})
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `{"data":{"iteratorState":{"blockNonce":1,"blockHash":"","blockRootHash":"","iteratorState":null}},"error":"","code":"successful"}`, w.Body.String())

		c, w = createPagedContext(http.MethodGet, "", "")
		RespondWithPage(c, nil, nil)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `{"data":null,"error":"","code":"successful"}`, w.Body.String())
	})
	t.Run("with iterator state", func(t *testing.T) {
		t.Parallel()

		c, w := createPagedContext(http.MethodGet, "", "")
		RespondWithPage(c, gin.H{"pairs": map[string]string{}, "blockInfo": api.BlockInfo{Nonce: 3}}, &checkpoint.Checkpoint{BlockNonce: 3, IteratorState: [][]byte{[]byte("s")}})
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `{"data":{"blockInfo":{"nonce":3},"iteratorState":{"blockNonce":3,"blockHash":"","blockRootHash":"","iteratorState":["cw=="]},"pairs":{}},"error":"","code":"successful"}`, w.Body.String())
	})
}

func TestRespondWithPageError(t *testing.T) {
	t.Parallel()

	for innerErr, expectedStatus := range map[error]int{
		checkpoint.ErrInvalidCheckpoint:                                                  http.StatusBadRequest,
		apiErrors.ErrBadUrlParams:                                                        http.StatusBadRequest,
		apiErrors.ErrValidation:                                                          http.StatusBadRequest,
		common.ErrInvalidIteratorState:                                                   http.StatusBadRequest,
		common.ErrIteratorNodeUnavailable:                                                http.StatusBadRequest,
		common.ErrTrieOperationsBusy:                                                     http.StatusTooManyRequests,
		errors.New("storage error"):                                                      http.StatusInternalServerError,
		errors.New("wrapped: " + "not matched"):                                          http.StatusInternalServerError,
		fmt.Errorf("node: %w", common.ErrIteratorNodeUnavailable):                        http.StatusBadRequest,
		fmt.Errorf("node: %w", common.ErrInvalidIteratorState):                           http.StatusBadRequest,
		fmt.Errorf("node: %w", common.ErrTrieOperationsBusy):                             http.StatusTooManyRequests,
		fmt.Errorf("%w: %w", apiErrors.ErrBadUrlParams, checkpoint.ErrInvalidCheckpoint): http.StatusBadRequest,
		fmt.Errorf("%w: %w", apiErrors.ErrValidation, &http.MaxBytesError{Limit: 10}):    http.StatusRequestEntityTooLarge,
	} {
		c, w := createPagedContext(http.MethodGet, "", "")
		RespondWithPageError(c, apiErrors.ErrGetKeyValuePairs, innerErr)
		require.Equal(t, expectedStatus, w.Code, innerErr.Error())
		require.Contains(t, w.Body.String(), apiErrors.ErrGetKeyValuePairs.Error()+": "+innerErr.Error())
	}

	c, w := createPagedContext(http.MethodGet, "", "")
	RespondWithPageError(c, apiErrors.ErrGetKeyValuePairs, errors.Join(errors.New("context"), common.ErrIteratorNodeUnavailable))
	require.Equal(t, http.StatusBadRequest, w.Code)

	// a body above the limit read while decoding (unknown length) is reported like one announced by its length
	c, w = createPagedContext(http.MethodPost, "", `{"numKeys": 1234567890}`)
	c.Request.Body = http.MaxBytesReader(w, c.Request.Body, 10)
	_, err := ExtractPagedRequest(c)
	var tooLarge *http.MaxBytesError
	require.ErrorAs(t, err, &tooLarge)
	RespondWithPageError(c, apiErrors.ErrGetKeyValuePairs, err)
	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestPagedHelpers_GetAndPostThroughServer(t *testing.T) {
	t.Parallel()

	cp := &checkpoint.Checkpoint{BlockNonce: 10, BlockRootHash: "bb", IteratorState: [][]byte{[]byte("state")}}
	encodedCp, err := checkpoint.Encode(cp)
	require.Nil(t, err)
	cpJSON, err := json.Marshal(cp)
	require.Nil(t, err)

	handler := func(c *gin.Context) {
		request, err := ExtractPagedRequest(c)
		if err != nil {
			RespondWithPageError(c, apiErrors.ErrGetKeyValuePairs, err)
			return
		}

		var next *checkpoint.Checkpoint
		if request.IteratorState != nil {
			next = &checkpoint.Checkpoint{BlockNonce: request.IteratorState.BlockNonce + 1, IteratorState: [][]byte{[]byte("next")}}
		}
		RespondWithPage(c, gin.H{
			"address":       c.Param("address"),
			"numKeys":       request.NumKeys,
			"blockRootHash": request.Options.BlockRootHash,
			"blockNonce":    request.Options.BlockNonce.Value,
		}, next)
	}
	router := gin.New()
	router.Use(middleware.NewTrieOperationsTransport(config.TrieOperationsConfig{}).MiddlewareHandlerFunc())
	router.Use(middleware.NewTrieOperationsThrottler(config.TrieOperationsConfig{}).MiddlewareHandlerFunc())
	router.GET("/address/:address/keys", handler)
	router.POST("/address/:address/keys", handler)
	server := httptest.NewServer(router)
	defer server.Close()

	type response struct {
		Data struct {
			Address       string                 `json:"address"`
			NumKeys       uint                   `json:"numKeys"`
			BlockRootHash []byte                 `json:"blockRootHash"`
			BlockNonce    uint64                 `json:"blockNonce"`
			IteratorState *checkpoint.Checkpoint `json:"iteratorState"`
		} `json:"data"`
		Error string            `json:"error"`
		Code  shared.ReturnCode `json:"code"`
	}
	do := func(method string, query string, body string) (int, response) {
		req, err := http.NewRequest(method, server.URL+"/address/alice/keys?"+query, strings.NewReader(body))
		require.Nil(t, err)
		resp, err := server.Client().Do(req)
		require.Nil(t, err)
		defer func() { _ = resp.Body.Close() }()
		parsed := response{}
		require.Nil(t, json.NewDecoder(resp.Body).Decode(&parsed))

		return resp.StatusCode, parsed
	}

	status, resp := do(http.MethodGet, "numKeys=3&iteratorState="+encodedCp+"&blockRootHash=bb", "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, shared.ReturnCodeSuccess, resp.Code)
	require.Equal(t, "alice", resp.Data.Address)
	require.Equal(t, uint(3), resp.Data.NumKeys)
	require.Equal(t, []byte{0xbb}, resp.Data.BlockRootHash)
	require.Equal(t, &checkpoint.Checkpoint{BlockNonce: 11, IteratorState: [][]byte{[]byte("next")}}, resp.Data.IteratorState)

	status, resp = do(http.MethodPost, "", `{"numKeys":4,"iteratorState":`+string(cpJSON)+`,"blockRootHash":"bb"}`)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, "alice", resp.Data.Address)
	require.Equal(t, uint(4), resp.Data.NumKeys)
	require.Equal(t, []byte{0xbb}, resp.Data.BlockRootHash)
	require.Equal(t, &checkpoint.Checkpoint{BlockNonce: 11, IteratorState: [][]byte{[]byte("next")}}, resp.Data.IteratorState)

	status, resp = do(http.MethodPost, "numKeys=3&blockNonce=1", `{"numKeys":4,"blockNonce":2}`)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, uint(4), resp.Data.NumKeys)
	require.Equal(t, uint64(2), resp.Data.BlockNonce)
	require.Nil(t, resp.Data.IteratorState)

	status, resp = do(http.MethodPost, "numKeys=3", "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, uint(3), resp.Data.NumKeys)
	require.Nil(t, resp.Data.IteratorState)

	status, resp = do(http.MethodGet, "numKeys=3", "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, uint(3), resp.Data.NumKeys)
	require.Nil(t, resp.Data.IteratorState)

	status, resp = do(http.MethodGet, "iteratorState=bm90IGpzb24", "")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, shared.ReturnCodeRequestError, resp.Code)
	require.Contains(t, resp.Error, checkpoint.ErrInvalidCheckpoint.Error())

	status, resp = do(http.MethodPost, "", `{"iteratorState":{"iteratorState":["%%"]}}`)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, shared.ReturnCodeRequestError, resp.Code)
	require.Contains(t, resp.Error, apiErrors.ErrValidation.Error())

	status, resp = do(http.MethodPost, "blockNonce=1", `{"blockHash":"dd"}`)
	require.Equal(t, http.StatusBadRequest, status)
	require.Contains(t, resp.Error, "only one block coordinate")
}
