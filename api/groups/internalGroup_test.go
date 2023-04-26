package groups_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
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

var (
	expectedRawBlockOutput = bytes.Repeat([]byte("1"), 10)
	expectedMetaBlock      = block.MetaBlock{
		Nonce: 15,
		Epoch: 15,
	}
	expectedShardBlock = block.Header{
		Nonce: 15,
		Round: 15,
	}
)

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

func TestInternalBlockGroup_getMetaBlockByNonce(t *testing.T) {
	t.Parallel()

	t.Run("empty nonce should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/metablock/by-nonce", nil, http.StatusNotFound, "")
	})
	t.Run("invalid nonce should error",
		testInternalGroupErrorScenario("/internal/raw/metablock/by-nonce/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockNonce)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/metablock/by-nonce/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/metablock/by-nonce/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawMetaBlockByRound(t *testing.T) {
	t.Parallel()

	t.Run("empty round should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/metablock/by-round", nil, http.StatusNotFound, "")
	})
	t.Run("invalid round should error",
		testInternalGroupErrorScenario("/internal/raw/metablock/by-round/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockRound)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/metablock/by-round/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/metablock/by-round/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawMetaBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/metablock/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/metablock/by-hash/dummyhash",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/metablock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawStartOfEpochMetaBlock(t *testing.T) {
	t.Parallel()

	t.Run("empty epoch should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/startofepoch/metablock/by-epoch/", nil, http.StatusNotFound, "")
	})
	t.Run("invalid epoch should error",
		testInternalGroupErrorScenario("/internal/raw/startofepoch/metablock/by-epoch/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidEpoch)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/startofepoch/metablock/by-epoch/1",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/startofepoch/metablock/by-epoch/1",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawShardBlockByNonce(t *testing.T) {
	t.Parallel()

	t.Run("empty nonce should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/shardblock/by-nonce", nil, http.StatusNotFound, "")
	})
	t.Run("invalid nonce should error",
		testInternalGroupErrorScenario("/internal/raw/shardblock/by-nonce/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockNonce)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/shardblock/by-nonce/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/shardblock/by-nonce/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawShardBlockByRound(t *testing.T) {
	t.Parallel()

	t.Run("empty round should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/shardblock/by-round", nil, http.StatusNotFound, "")
	})
	t.Run("invalid round should error",
		testInternalGroupErrorScenario("/internal/raw/shardblock/by-round/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockRound)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/shardblock/by-round/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/shardblock/by-round/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawShardBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/shardblock/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/shardblock/by-hash/dummyhash",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/shardblock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getRawMiniBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/miniblock/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("empty epoch should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/raw/miniblock/by-hash/aaa/epoch", nil, http.StatusNotFound, "")
	})
	t.Run("invalid epoch should error",
		testInternalGroupErrorScenario("/internal/raw/miniblock/by-hash/aaaa/epoch/not-uint", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidEpoch)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMiniBlockByHashCalled: func(format common.ApiOutputFormat, txHash string, epoch uint32) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/raw/miniblock/by-hash/aaaa/epoch/1",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMiniBlockByHashCalled: func(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error) {
				return expectedRawBlockOutput, nil
			},
		}

		response := &rawMiniBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/raw/miniblock/by-hash/aaaa/epoch/1",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedRawBlockOutput, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONMetaBlockByNonce(t *testing.T) {
	t.Parallel()

	t.Run("empty nonce should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/metablock/by-nonce", nil, http.StatusNotFound, "")
	})
	t.Run("invalid nonce should error",
		testInternalGroupErrorScenario("/internal/json/metablock/by-nonce/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockNonce)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/metablock/by-nonce/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedMetaBlock, nil
			},
		}

		response := &internalMetaBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/metablock/by-nonce/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedMetaBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONMetaBlockByRound(t *testing.T) {
	t.Parallel()

	t.Run("empty round should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/metablock/by-round", nil, http.StatusNotFound, "")
	})
	t.Run("invalid round should error",
		testInternalGroupErrorScenario("/internal/json/metablock/by-round/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockRound)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/metablock/by-round/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedMetaBlock, nil
			},
		}

		response := &internalMetaBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/metablock/by-round/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedMetaBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONMetaBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/metablock/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/metablock/by-hash/dummyhash",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return expectedMetaBlock, nil
			},
		}

		response := &internalMetaBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/metablock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedMetaBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONStartOfEpochMetaBlock(t *testing.T) {
	t.Parallel()

	t.Run("empty epoch should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/startofepoch/metablock/by-epoch/", nil, http.StatusNotFound, "")
	})
	t.Run("invalid epoch should error",
		testInternalGroupErrorScenario("/internal/json/startofepoch/metablock/by-epoch/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidEpoch)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/startofepoch/metablock/by-epoch/1",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochMetaBlockCalled: func(_ common.ApiOutputFormat, epoch uint32) (interface{}, error) {
				return expectedMetaBlock, nil
			},
		}

		response := &internalMetaBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/startofepoch/metablock/by-epoch/1",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedMetaBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONShardBlockByNonce(t *testing.T) {
	t.Parallel()

	t.Run("empty nonce should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/shardblock/by-nonce", nil, http.StatusNotFound, "")
	})
	t.Run("invalid nonce should error",
		testInternalGroupErrorScenario("/internal/json/shardblock/by-nonce/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockNonce)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/shardblock/by-nonce/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedShardBlock, nil
			},
		}

		response := &internalShardBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/shardblock/by-nonce/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedShardBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONShardBlockByRound(t *testing.T) {
	t.Parallel()

	t.Run("empty round should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/shardblock/by-round", nil, http.StatusNotFound, "")
	})
	t.Run("invalid round should error",
		testInternalGroupErrorScenario("/internal/json/shardblock/by-round/invalid", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidBlockRound)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/shardblock/by-round/15",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
				return expectedShardBlock, nil
			},
		}

		response := &internalShardBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/shardblock/by-round/15",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedShardBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONShardBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/shardblock/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/shardblock/by-hash/dummyhash",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
				return expectedShardBlock, nil
			},
		}

		response := &internalShardBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/shardblock/by-hash/d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedShardBlock, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONMiniBlockByHash(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/miniblock/by-hash", nil, http.StatusNotFound, "")
	})
	t.Run("empty epoch should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/miniblock/by-hash/aaa/epoch", nil, http.StatusNotFound, "")
	})
	t.Run("invalid epoch should error",
		testInternalGroupErrorScenario("/internal/json/miniblock/by-hash/aaaa/epoch/not-uint", nil,
			formatExpectedErr(apiErrors.ErrGetBlock, apiErrors.ErrInvalidEpoch)))
	t.Run("facade error should error", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMiniBlockByHashCalled: func(format common.ApiOutputFormat, txHash string, epoch uint32) (interface{}, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/miniblock/by-hash/aaaa/epoch/1",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetBlock, expectedErr),
		)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalMiniBlockByHashCalled: func(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error) {
				return block.MiniBlock{}, nil
			},
		}

		response := &internalMiniBlockResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/miniblock/by-hash/aaaa/epoch/1",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, block.MiniBlock{}, response.Data.Block)
	})
}

func TestInternalBlockGroup_getJSONStartOfEpochValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("empty epoch should error", func(t *testing.T) {
		t.Parallel()

		testInternalGroup(t, &mock.FacadeStub{}, "/internal/json/startofepoch/validators/by-epoch", nil, http.StatusNotFound, "")
	})
	t.Run("invalid epoch should error",
		testInternalGroupErrorScenario("/internal/json/startofepoch/validators/by-epoch/not-uint", nil,
			formatExpectedErr(apiErrors.ErrGetValidatorsInfo, apiErrors.ErrInvalidEpoch)))
	t.Run("facade error should fail", func(t *testing.T) {
		t.Parallel()

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return nil, expectedErr
			},
		}

		testInternalGroup(
			t,
			facade,
			"/internal/json/startofepoch/validators/by-epoch/1",
			nil,
			http.StatusInternalServerError,
			formatExpectedErr(apiErrors.ErrGetValidatorsInfo, expectedErr),
		)
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

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return expectedOutput, nil
			},
		}

		response := &internalValidatorsInfoResponse{}
		loadInternalBlockGroupResponse(
			t,
			facade,
			"/internal/json/startofepoch/validators/by-epoch/1",
			"GET",
			nil,
			response,
		)
		assert.Equal(t, expectedOutput, response.Data.ValidatorsInfo)
	})
}

func TestInternalBlockGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	blockGroup, _ := groups.NewInternalBlockGroup(nil)
	require.True(t, blockGroup.IsInterfaceNil())

	blockGroup, _ = groups.NewInternalBlockGroup(&mock.FacadeStub{})
	require.False(t, blockGroup.IsInterfaceNil())
}

func TestInternalBlockGroup_UpdateFacadeStub(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		blockGroup, err := groups.NewInternalBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = blockGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		blockGroup, err := groups.NewInternalBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = blockGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
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

		facade := &mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return expectedOutput, nil
			},
		}

		blockGroup, err := groups.NewInternalBlockGroup(facade)
		require.NoError(t, err)

		ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

		req, _ := http.NewRequest("GET", "/internal/json/startofepoch/validators/by-epoch/1", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := internalValidatorsInfoResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusOK, resp.Code)

		assert.Equal(t, expectedOutput, response.Data.ValidatorsInfo)

		newFacade := &mock.FacadeStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return nil, expectedErr
			},
		}
		err = blockGroup.UpdateFacade(newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/internal/json/startofepoch/validators/by-epoch/1", nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response = internalValidatorsInfoResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
}

func loadInternalBlockGroupResponse(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	method string,
	body io.Reader,
	destination interface{},
) {
	blockGroup, err := groups.NewInternalBlockGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest(method, url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	loadResponse(resp.Body, destination)
}

func testInternalGroupErrorScenario(url string, body io.Reader, expectedErr string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		testInternalGroup(
			t,
			&mock.FacadeStub{},
			url,
			body,
			http.StatusBadRequest,
			expectedErr,
		)
	}
}

func testInternalGroup(
	t *testing.T,
	facade shared.FacadeHandler,
	url string,
	body io.Reader,
	expectedRespCode int,
	expectedRespError string,
) {
	internalGroup, err := groups.NewInternalBlockGroup(facade)
	require.NoError(t, err)

	ws := startWebServer(internalGroup, "internal", getInternalBlockRoutesConfig())

	req, _ := http.NewRequest("GET", url, body)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, expectedRespCode, resp.Code)
	assert.True(t, strings.Contains(response.Error, expectedRespError))
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
