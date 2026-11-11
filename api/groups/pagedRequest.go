package groups

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	customErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
)

const (
	urlParamNumKeys       = "numKeys"
	urlParamIteratorState = "iteratorState"
)

// PagedRequest holds the parameters of a paginated trie walk
type PagedRequest struct {
	Options       api.AccountQueryOptions
	NumKeys       uint
	IteratorState *checkpoint.Checkpoint
}

// PagedRequestBody is the JSON body accepted by a paginated trie walk. Every field is optional and, when
// present, overrides the query parameter with the same name; a null value or an empty hash string counts as absent
type PagedRequestBody struct {
	NumKeys        *uint                  `json:"numKeys"`
	IteratorState  *checkpoint.Checkpoint `json:"iteratorState"`
	OnFinalBlock   *bool                  `json:"onFinalBlock"`
	OnStartOfEpoch *uint32                `json:"onStartOfEpoch"`
	BlockNonce     *uint64                `json:"blockNonce"`
	BlockHash      string                 `json:"blockHash"`
	BlockRootHash  string                 `json:"blockRootHash"`
	HintEpoch      *uint32                `json:"hintEpoch"`
}

// ExtractPagedRequest reads the parameters of a paginated trie walk from the query and, for a POST request,
// from the JSON body, which overrides the query. The iterator state is the checkpoint object in the body or its
// base64url encoding in the query
func ExtractPagedRequest(c *gin.Context) (PagedRequest, error) {
	request, err := parsePagedQuery(c)
	if err != nil {
		return PagedRequest{}, fmt.Errorf("%w: %w", customErrors.ErrBadUrlParams, err)
	}

	if c.Request.Method == http.MethodPost {
		err = applyPagedBody(c, &request)
		if err != nil {
			return PagedRequest{}, fmt.Errorf("%w: %w", customErrors.ErrValidation, err)
		}
	}

	err = checkAccountQueryOptions(request.Options)
	if err != nil {
		return PagedRequest{}, fmt.Errorf("%w: %v", customErrors.ErrBadUrlParams, err)
	}

	return request, nil
}

func parsePagedQuery(c *gin.Context) (PagedRequest, error) {
	options, err := parseAccountQueryOptions(c)
	if err != nil {
		return PagedRequest{}, err
	}

	numKeys, err := parseUintUrlParam(c, urlParamNumKeys)
	if err != nil {
		return PagedRequest{}, err
	}

	request := PagedRequest{Options: options, NumKeys: numKeys}
	encoded := c.Request.URL.Query().Get(urlParamIteratorState)
	if encoded == "" {
		return request, nil
	}

	request.IteratorState, err = checkpoint.Decode(encoded)
	if err != nil {
		return PagedRequest{}, err
	}

	return request, nil
}

func applyPagedBody(c *gin.Context, request *PagedRequest) error {
	if c.Request.Body == nil {
		return nil
	}

	body := PagedRequestBody{}
	err := c.ShouldBindJSON(&body)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}

	if body.NumKeys != nil {
		request.NumKeys = *body.NumKeys
	}
	if body.IteratorState != nil {
		request.IteratorState = body.IteratorState
	}
	if body.OnFinalBlock != nil {
		request.Options.OnFinalBlock = *body.OnFinalBlock
	}
	if body.OnStartOfEpoch != nil {
		request.Options.OnStartOfEpoch = core.OptionalUint32{Value: *body.OnStartOfEpoch, HasValue: true}
	}
	if body.BlockNonce != nil {
		request.Options.BlockNonce = core.OptionalUint64{Value: *body.BlockNonce, HasValue: true}
	}
	if body.HintEpoch != nil {
		request.Options.HintEpoch = core.OptionalUint32{Value: *body.HintEpoch, HasValue: true}
	}
	if body.BlockHash != "" {
		request.Options.BlockHash, err = hex.DecodeString(body.BlockHash)
		if err != nil {
			return fmt.Errorf("blockHash: %w", err)
		}
	}
	if body.BlockRootHash != "" {
		request.Options.BlockRootHash, err = hex.DecodeString(body.BlockRootHash)
		if err != nil {
			return fmt.Errorf("blockRootHash: %w", err)
		}
	}

	return nil
}

// RespondWithPage responds with the data of a page; the iterator state is added only when the walk has more
// to return
func RespondWithPage(c *gin.Context, data gin.H, iteratorState *checkpoint.Checkpoint) {
	if iteratorState != nil {
		if data == nil {
			data = gin.H{}
		}
		data[urlParamIteratorState] = iteratorState
	}

	shared.RespondWithSuccess(c, data)
}

// RespondWithPageError responds to a failed paginated request: a body above the limit is too large, a request that
// can not be parsed or an iterator state that can not be used is a bad request, any other error is an internal
// error (or system busy)
func RespondWithPageError(c *gin.Context, err error, innerErr error) {
	var tooLarge *http.MaxBytesError
	if errors.As(innerErr, &tooLarge) {
		shared.RespondWith(c, http.StatusRequestEntityTooLarge, nil, fmt.Sprintf("%s: %s", err.Error(), innerErr.Error()), shared.ReturnCodeRequestError)
		return
	}
	if errors.Is(innerErr, customErrors.ErrBadUrlParams) ||
		errors.Is(innerErr, customErrors.ErrValidation) ||
		errors.Is(innerErr, customErrors.ErrEmptyAddress) ||
		errors.Is(innerErr, checkpoint.ErrInvalidCheckpoint) ||
		errors.Is(innerErr, common.ErrInvalidIteratorState) ||
		errors.Is(innerErr, common.ErrIteratorNodeUnavailable) {
		shared.RespondWithValidationError(c, err, innerErr)
		return
	}

	shared.RespondWithInternalError(c, err, innerErr)
}
