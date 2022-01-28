package groups

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/gin-gonic/gin"
)

const (
	getRawMetaBlockByNoncePath       = "/raw/metablock/by-nonce/:nonce"
	getRawMetaBlockByHashPath        = "/raw/metablock/by-hash/:hash"
	getRawMetaBlockByRoundPath       = "/raw/metablock/by-round/:round"
	getRawShardBlockByNoncePath      = "/raw/shardblock/by-nonce/:nonce"
	getRawShardBlockByHashPath       = "/raw/shardblock/by-hash/:hash"
	getRawShardBlockByRoundPath      = "/raw/shardblock/by-round/:round"
	getInternalMetaBlockByNoncePath  = "/json/metablock/by-nonce/:nonce"
	getInternalMetaBlockByHashPath   = "/json/metablock/by-hash/:hash"
	getInternalMetaBlockByRoundPath  = "/json/metablock/by-round/:round"
	getInternalShardBlockByNoncePath = "/json/shardblock/by-nonce/:nonce"
	getInternalShardBlockByHashPath  = "/json/shardblock/by-hash/:hash"
	getInternalShardBlockByRoundPath = "/json/shardblock/by-round/:round"
	getRawMiniBlockByHashPath        = "/raw/miniblock/by-hash/:hash"
	getInternalMiniBlockByHashPath   = "/json/miniblock/by-hash/:hash"
)

// internalBlockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type internalBlockFacadeHandler interface {
	GetInternalShardBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHash(format common.OutportFormat, hash string) (interface{}, error)
	GetInternalShardBlockByRound(format common.OutportFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHash(format common.OutportFormat, hash string) (interface{}, error)
	GetInternalMetaBlockByRound(format common.OutportFormat, round uint64) (interface{}, error)
	GetInternalMiniBlockByHash(format common.OutportFormat, hash string) (interface{}, error)
	IsInterfaceNil() bool
}

type internalBlockGroup struct {
	*baseGroup
	facade    internalBlockFacadeHandler
	mutFacade sync.RWMutex
}

// NewInternalBlockGroup returns a new instance of rawBlockGroup
func NewInternalBlockGroup(facade internalBlockFacadeHandler) (*internalBlockGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for internal block group", errors.ErrNilFacadeHandler)
	}

	ib := &internalBlockGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getRawMetaBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: ib.getRawMetaBlockByNonce,
		},
		{
			Path:    getRawMetaBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getRawMetaBlockByHash,
		},
		{
			Path:    getRawMetaBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: ib.getRawMetaBlockByRound,
		},
		{
			Path:    getRawShardBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: ib.getRawShardBlockByNonce,
		},
		{
			Path:    getRawShardBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getRawShardBlockByHash,
		},
		{
			Path:    getRawShardBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: ib.getRawShardBlockByRound,
		},
		{
			Path:    getInternalMetaBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: ib.getInternalMetaBlockByNonce,
		},
		{
			Path:    getInternalMetaBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getInternalMetaBlockByHash,
		},
		{
			Path:    getInternalMetaBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: ib.getInternalMetaBlockByRound,
		},
		{
			Path:    getInternalShardBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: ib.getInternalShardBlockByNonce,
		},
		{
			Path:    getInternalShardBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getInternalShardBlockByHash,
		},
		{
			Path:    getInternalShardBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: ib.getInternalShardBlockByRound,
		},
		{
			Path:    getRawMiniBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getRawMiniBlockByHash,
		},
		{
			Path:    getInternalMiniBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getInternalMiniBlockByHash,
		},
	}
	ib.endpoints = endpoints

	return ib, nil
}

func (ib *internalBlockGroup) getRawMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalMetaBlockByNonce(common.Proto, nonce)
	log.Debug(fmt.Sprintf("GetInternalMetaBlockByNonce with proto took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getRawMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalMetaBlockByHash(common.Proto, hash)
	log.Debug(fmt.Sprintf("GetInternalMetaBlockByHash with proto took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getRawMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalMetaBlockByRound(common.Proto, round)
	log.Debug(fmt.Sprintf("GetInternalMetaBlockByRound with proto took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getRawShardBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalShardBlockByNonce(common.Proto, nonce)
	log.Debug(fmt.Sprintf("GetInternalShardBlockByNonce with proto took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getRawShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalShardBlockByHash(common.Proto, hash)
	log.Debug(fmt.Sprintf("GetInternalShardBlockByHash with proto took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getRawShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalShardBlockByRound(common.Proto, round)
	log.Debug(fmt.Sprintf("GetInternalShardBlockByRound with proto took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getInternalMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalMetaBlockByNonce(common.Internal, nonce)
	log.Debug(fmt.Sprintf("GetInternalMetaBlockByNonce took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getInternalMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalMetaBlockByHash(common.Internal, hash)
	log.Debug(fmt.Sprintf("GetInternalMetaBlockByHash took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getInternalMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalMetaBlockByRound(common.Internal, round)
	log.Debug(fmt.Sprintf("GetInternalMetaBlockByRound took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getInternalShardBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalShardBlockByNonce(common.Internal, nonce)
	log.Debug(fmt.Sprintf("GetInternalShardBlockByNonce took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getInternalShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalShardBlockByHash(common.Internal, hash)
	log.Debug(fmt.Sprintf("GetInternalShardBlockByHash took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getInternalShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalShardBlockByRound(common.Internal, round)
	log.Debug(fmt.Sprintf("GetInternalShardBlockByRound took %s", time.Since(start)))
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

func (ib *internalBlockGroup) getRawMiniBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	miniBlock, err := ib.getFacade().GetInternalMiniBlockByHash(common.Proto, hash)
	log.Debug(fmt.Sprintf("GetInternalMiniBlockByHash with proto took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"miniblock": miniBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getInternalMiniBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	start := time.Now()
	miniBlock, err := ib.getFacade().GetInternalMiniBlockByHash(common.Internal, hash)
	log.Debug(fmt.Sprintf("GetInternalMiniBlockByHash took %s", time.Since(start)))
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

	shared.RespondWith(c, http.StatusOK, gin.H{"miniblock": miniBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getFacade() internalBlockFacadeHandler {
	ib.mutFacade.RLock()
	defer ib.mutFacade.RUnlock()

	return ib.facade
}

// UpdateFacade will update the facade
func (ib *internalBlockGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(internalBlockFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	ib.mutFacade.Lock()
	ib.facade = castFacade
	ib.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ib *internalBlockGroup) IsInterfaceNil() bool {
	return ib == nil
}
