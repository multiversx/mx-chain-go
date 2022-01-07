package groups

import (
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
	getRawMetaBlockByNoncePath   = "/raw/metablock/by-nonce/:nonce"
	getRawMetaBlockByHashPath    = "/raw/metablock/by-hash/:hash"
	getRawMetaBlockByRoundPath   = "/raw/metablock/by-round/:round"
	getRawShardBlockByNoncePath  = "/raw/shardblock/by-nonce/:nonce"
	getRawShardBlockByHashPath   = "/raw/shardblock/by-hash/:hash"
	getRawShardBlockByRoundPath  = "/raw/shardblock/by-round/:round"
	getJsonMetaBlockByNoncePath  = "/json/metablock/by-nonce/:nonce"
	getJsonMetaBlockByHashPath   = "/json/metablock/by-hash/:hash"
	getJsonMetaBlockByRoundPath  = "/json/metablock/by-round/:round"
	getJsonShardBlockByNoncePath = "/json/shardblock/by-nonce/:nonce"
	getJsonShardBlockByHashPath  = "/json/shardblock/by-hash/:hash"
	getJsonShardBlockByRoundPath = "/json/shardblock/by-round/:round"
	getRawMiniBlockByHashPath    = "/miniblock/by-hash/:hash"
)

// TODO: comments update

// rawBlockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type rawBlockFacadeHandler interface {
	GetRawMetaBlockByHash(hash string) ([]byte, error)
	GetRawMetaBlockByNonce(nonce uint64) ([]byte, error)
	GetRawMetaBlockByRound(round uint64) ([]byte, error)
	GetRawShardBlockByHash(hash string) ([]byte, error)
	GetRawShardBlockByNonce(nonce uint64) ([]byte, error)
	GetRawShardBlockByRound(round uint64) ([]byte, error)
	GetJsonMetaBlockByHash(hash string) (*block.MetaBlock, error)
	GetJsonMetaBlockByNonce(nonce uint64) (*block.MetaBlock, error)
	GetJsonMetaBlockByRound(round uint64) (*block.MetaBlock, error)
	GetJsonShardBlockByHash(hash string) (*block.Header, error)
	GetJsonShardBlockByNonce(nonce uint64) (*block.Header, error)
	GetJsonShardBlockByRound(round uint64) (*block.Header, error)
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
		// JSON
		{
			Path:    getJsonMetaBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: rb.getJsonMetaBlockByNonce,
		},
		{
			Path:    getJsonMetaBlockByHashPath,
			Method:  http.MethodGet,
			Handler: rb.getJsonMetaBlockByHash,
		},
		{
			Path:    getJsonMetaBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: rb.getJsonMetaBlockByRound,
		},
		{
			Path:    getJsonShardBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: rb.getJsonShardBlockByNonce,
		},
		{
			Path:    getJsonShardBlockByHashPath,
			Method:  http.MethodGet,
			Handler: rb.getJsonShardBlockByHash,
		},
		{
			Path:    getJsonShardBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: rb.getJsonShardBlockByRound,
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

// ---- Raw Blocks

func (rb *rawBlockGroup) getRawMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawMetaBlockByNonce(nonce)
	log.Debug(fmt.Sprintf("GetRawMetaBlockByNonce took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getRawMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawMetaBlockByHash(hash)
	log.Debug(fmt.Sprintf("GetRawMetaBlockByHash took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getRawMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawMetaBlockByRound(round)
	log.Debug(fmt.Sprintf("GetRawMetaBlockByRound took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
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

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawShardBlockByNonce(nonce)
	log.Debug(fmt.Sprintf("GetRawShardBlockByNonce took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getRawShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawShardBlockByHash(hash)
	log.Debug(fmt.Sprintf("GetRawShardBlockByHash took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getRawShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := rb.getFacade().GetRawShardBlockByRound(round)
	log.Debug(fmt.Sprintf("GetRawShardBlockByRound took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

// ---- Json Blocks

func (rb *rawBlockGroup) getJsonMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := rb.getFacade().GetJsonMetaBlockByNonce(nonce)
	log.Debug(fmt.Sprintf("GetJsonMetaBlockByNonce took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getJsonMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	block, err := rb.getFacade().GetJsonMetaBlockByHash(hash)
	log.Debug(fmt.Sprintf("GetJsonMetaBlockByHash took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getJsonMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	block, err := rb.getFacade().GetJsonMetaBlockByRound(round)
	log.Debug(fmt.Sprintf("GetJsonMetaBlockByRound took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

// --------------------------- shard block -----------------------------

func (rb *rawBlockGroup) getJsonShardBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := rb.getFacade().GetJsonShardBlockByNonce(nonce)
	log.Debug(fmt.Sprintf("GetJsonShardBlockByNonce took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getJsonShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	block, err := rb.getFacade().GetJsonShardBlockByHash(hash)
	log.Debug(fmt.Sprintf("GetJsonShardBlockByHash took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (rb *rawBlockGroup) getJsonShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	block, err := rb.getFacade().GetJsonShardBlockByRound(round)
	log.Debug(fmt.Sprintf("GetJsonShardBlockByRound took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

// ---- Raw MiniBlock

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
	log.Debug(fmt.Sprintf("GetRawMiniBlockByHash took %s", time.Since(start)))
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
