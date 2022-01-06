package groups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	getRawMetaBlockByNoncePath  = "/metablock/by-nonce/:nonce"
	getRawMetaBlockByHashPath   = "/metablock/by-hash/:hash"
	getRawMetaBlockByRoundPath  = "/metablock/by-round/:round"
	getRawShardBlockByNoncePath = "/shardblock/by-nonce/:nonce"
	getRawShardBlockByHashPath  = "/shardblock/by-hash/:hash"
	getRawShardBlockByRoundPath = "/shardblock/by-round/:round"
	getRawMiniBlockByHashPath   = "/miniblock/by-hash/:hash"
)

// TODO: comments update

// rawBlockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type rawBlockFacadeHandler interface {
	GetRawMetaBlockByHash(hash string, asJson bool) ([]byte, error)
	GetRawMetaBlockByNonce(nonce uint64, asJson bool) ([]byte, error)
	GetRawMetaBlockByRound(round uint64, asJson bool) ([]byte, error)
	GetRawShardBlockByHash(hash string, asJson bool) ([]byte, error)
	GetRawShardBlockByNonce(nonce uint64, asJson bool) ([]byte, error)
	GetRawShardBlockByRound(round uint64, asJson bool) ([]byte, error)
	GetRawMiniBlockByHash(hash string) ([]byte, error)
	IsInterfaceNil() bool
}

type rawBlockGroup struct {
	*baseGroup
	facade    rawBlockFacadeHandler
	mutFacade sync.RWMutex
}

// NewBlockGroup returns a new instance of blockGroup
func NewRawBlockGroup(facade rawBlockFacadeHandler) (*rawBlockGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for raw block group", errors.ErrNilFacadeHandler)
	}

	rb := &rawBlockGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getRawMetaBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: rb.getRawMetaBlockByNonce,
		},
		{
			Path:    getRawMetaBlockByHashPath,
			Method:  http.MethodGet,
			Handler: rb.getRawMetaBlockByHash,
		},
		{
			Path:    getRawMetaBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: rb.getRawMetaBlockByRound,
		},
		{
			Path:    getRawShardBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: rb.getRawShardBlockByNonce,
		},
		{
			Path:    getRawShardBlockByHashPath,
			Method:  http.MethodGet,
			Handler: rb.getRawShardBlockByHash,
		},
		{
			Path:    getRawShardBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: rb.getRawShardBlockByRound,
		},
		{
			Path:    getRawMiniBlockByHashPath,
			Method:  http.MethodGet,
			Handler: rb.getRawMiniBlockByHash,
		},
	}
	rb.endpoints = endpoints

	return rb, nil
}

func (rb *rawBlockGroup) getRawMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	asJson, err := getQueryParamAsJSON(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawMetaBlockByNonce(nonce, asJson)
	log.Debug(fmt.Sprintf("GetBlockByNonce took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	if asJson {
		blockHeader := &block.MetaBlock{}
		err = json.Unmarshal(rawBlock, blockHeader)
		shared.RespondWith(c, http.StatusOK, gin.H{"block": blockHeader}, "", shared.ReturnCodeSuccess)
	} else {
		shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
	}
}

func (rb *rawBlockGroup) getRawMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	asJson, err := getQueryParamAsJSON(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawMetaBlockByHash(hash, asJson)
	log.Debug(fmt.Sprintf("GetBlockByHash took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	if asJson {
		blockHeader := &block.MetaBlock{}
		err = json.Unmarshal(rawBlock, blockHeader)
		shared.RespondWith(c, http.StatusOK, gin.H{"block": blockHeader}, "", shared.ReturnCodeSuccess)
	} else {
		shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
	}
}

func (rb *rawBlockGroup) getRawMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	asJson, err := getQueryParamAsJSON(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawMetaBlockByRound(round, asJson)
	log.Debug(fmt.Sprintf("GetBlockByRound took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	if asJson {
		blockHeader := &block.MetaBlock{}
		err = json.Unmarshal(rawBlock, blockHeader)
		shared.RespondWith(c, http.StatusOK, gin.H{"block": blockHeader}, "", shared.ReturnCodeSuccess)
	} else {
		shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
	}
}

// --------------------------- shard block -----------------------------

func (rb *rawBlockGroup) getRawShardBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	asJson, err := getQueryParamAsJSON(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawShardBlockByNonce(nonce, asJson)
	log.Debug(fmt.Sprintf("GetBlockByNonce took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	if asJson {
		blockHeader := &block.Header{}
		err = json.Unmarshal(rawBlock, blockHeader)
		shared.RespondWith(c, http.StatusOK, gin.H{"block": blockHeader}, "", shared.ReturnCodeSuccess)
	} else {
		shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
	}
}

func (rb *rawBlockGroup) getRawShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	asJson, err := getQueryParamAsJSON(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawShardBlockByHash(hash, asJson)
	log.Debug(fmt.Sprintf("GetBlockByHash took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	if asJson {
		blockHeader := &block.Header{}
		err = json.Unmarshal(rawBlock, blockHeader)
		shared.RespondWith(c, http.StatusOK, gin.H{"block": blockHeader}, "", shared.ReturnCodeSuccess)
	} else {
		shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
	}
}

func (rb *rawBlockGroup) getRawShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	asJson, err := getQueryParamAsJSON(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawShardBlockByRound(round, asJson)
	log.Debug(fmt.Sprintf("GetBlockByRound took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	if asJson {
		blockHeader := &block.Header{}
		err = json.Unmarshal(rawBlock, blockHeader)
		shared.RespondWith(c, http.StatusOK, gin.H{"block": blockHeader}, "", shared.ReturnCodeSuccess)
	} else {
		shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
	}
}

func (rb *rawBlockGroup) getRawMiniBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	miniBlock, err := rb.getFacade().GetRawMiniBlockByHash(hash)
	log.Debug(fmt.Sprintf("GetBlockByHash took %s", time.Since(start)))
	if err != nil {
		shared.RespondWith(
			c,
			http.StatusInternalServerError,
			nil,
			fmt.Sprintf("%s: %s", errors.ErrGetBlock.Error(), err.Error()),
			shared.ReturnCodeInternalError,
		)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": miniBlock}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getFacade() rawBlockFacadeHandler {
	rb.mutFacade.RLock()
	defer rb.mutFacade.RUnlock()

	return rb.facade
}

func getQueryParamAsJSON(c *gin.Context) (bool, error) {
	asJsonStr := c.Request.URL.Query().Get("asJSON")
	if asJsonStr == "" {
		return false, nil
	}

	return strconv.ParseBool(asJsonStr)
}

// UpdateFacade will update the facade
func (rb *rawBlockGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(rawBlockFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	rb.mutFacade.Lock()
	rb.facade = castFacade
	rb.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rb *rawBlockGroup) IsInterfaceNil() bool {
	return rb == nil
}
