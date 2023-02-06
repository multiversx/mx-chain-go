package groups

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/api/shared/logging"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

const (
	getRawMetaBlockByNoncePath            = "/raw/metablock/by-nonce/:nonce"
	getRawMetaBlockByHashPath             = "/raw/metablock/by-hash/:hash"
	getRawMetaBlockByRoundPath            = "/raw/metablock/by-round/:round"
	getRawStartOfEpochMetaBlockPath       = "/raw/startofepoch/metablock/by-epoch/:epoch"
	getRawShardBlockByNoncePath           = "/raw/shardblock/by-nonce/:nonce"
	getRawShardBlockByHashPath            = "/raw/shardblock/by-hash/:hash"
	getRawShardBlockByRoundPath           = "/raw/shardblock/by-round/:round"
	getJSONMetaBlockByNoncePath           = "/json/metablock/by-nonce/:nonce"
	getJSONMetaBlockByHashPath            = "/json/metablock/by-hash/:hash"
	getJSONMetaBlockByRoundPath           = "/json/metablock/by-round/:round"
	getJSONStartOfEpochMetaBlockPath      = "/json/startofepoch/metablock/by-epoch/:epoch"
	getJSONStartOfEpochValidatorsInfoPath = "/json/startofepoch/validators/by-epoch/:epoch"
	getJSONShardBlockByNoncePath          = "/json/shardblock/by-nonce/:nonce"
	getJSONShardBlockByHashPath           = "/json/shardblock/by-hash/:hash"
	getJSONShardBlockByRoundPath          = "/json/shardblock/by-round/:round"
	getRawMiniBlockByHashPath             = "/raw/miniblock/by-hash/:hash/epoch/:epoch"
	getJSONMiniBlockByHashPath            = "/json/miniblock/by-hash/:hash/epoch/:epoch"
)

// internalBlockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type internalBlockFacadeHandler interface {
	GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMiniBlockByHash(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error)
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
			Path:    getRawStartOfEpochMetaBlockPath,
			Method:  http.MethodGet,
			Handler: ib.getRawStartOfEpochMetaBlock,
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
			Path:    getJSONMetaBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: ib.getJSONMetaBlockByNonce,
		},
		{
			Path:    getJSONMetaBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONMetaBlockByHash,
		},
		{
			Path:    getJSONMetaBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONMetaBlockByRound,
		},
		{
			Path:    getJSONStartOfEpochMetaBlockPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONStartOfEpochMetaBlock,
		},
		{
			Path:    getJSONShardBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: ib.getJSONShardBlockByNonce,
		},
		{
			Path:    getJSONShardBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONShardBlockByHash,
		},
		{
			Path:    getJSONShardBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONShardBlockByRound,
		},
		{
			Path:    getRawMiniBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getRawMiniBlockByHash,
		},
		{
			Path:    getJSONMiniBlockByHashPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONMiniBlockByHash,
		},
		{
			Path:    getJSONStartOfEpochValidatorsInfoPath,
			Method:  http.MethodGet,
			Handler: ib.getJSONStartOfEpochValidatorsInfo,
		},
	}
	ib.endpoints = endpoints

	return ib, nil
}

func (ib *internalBlockGroup) getRawMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockNonce)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalMetaBlockByNonce(common.ApiOutputFormatProto, nonce)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMetaBlockByNonce with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalMetaBlockByHash(common.ApiOutputFormatProto, hash)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMetaBlockByHash with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockRound)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalMetaBlockByRound(common.ApiOutputFormatProto, round)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMetaBlockByRound with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawStartOfEpochMetaBlock(c *gin.Context) {
	epoch, err := getQueryParamEpoch(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidEpoch)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalStartOfEpochMetaBlock(common.ApiOutputFormatProto, epoch)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalStartOfEpochMetaBlock with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawShardBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockNonce)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalShardBlockByNonce(common.ApiOutputFormatProto, nonce)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalShardBlockByNonce with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalShardBlockByHash(common.ApiOutputFormatProto, hash)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalShardBlockByHash with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockRound)
		return
	}

	start := time.Now()
	rawBlock, err := ib.getFacade().GetInternalShardBlockByRound(common.ApiOutputFormatProto, round)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalShardBlockByRound with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": rawBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockNonce)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalMetaBlockByNonce(common.ApiOutputFormatJSON, nonce)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMetaBlockByNonce with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalMetaBlockByHash(common.ApiOutputFormatJSON, hash)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMetaBlockByHash with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockRound)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalMetaBlockByRound(common.ApiOutputFormatJSON, round)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMetaBlockByRound with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONStartOfEpochMetaBlock(c *gin.Context) {
	epoch, err := getQueryParamEpoch(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidEpoch)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalStartOfEpochMetaBlock(common.ApiOutputFormatJSON, epoch)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalStartOfEpochMetaBlock with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONShardBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockNonce)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalShardBlockByNonce(common.ApiOutputFormatJSON, nonce)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalShardBlockByNonce with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONShardBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalShardBlockByHash(common.ApiOutputFormatJSON, hash)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalShardBlockByHash with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONShardBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidBlockRound)
		return
	}

	start := time.Now()
	block, err := ib.getFacade().GetInternalShardBlockByRound(common.ApiOutputFormatJSON, round)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalShardBlockByRound with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getRawMiniBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	epoch, err := getQueryParamEpoch(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidEpoch)
		return
	}

	start := time.Now()
	miniBlock, err := ib.getFacade().GetInternalMiniBlockByHash(common.ApiOutputFormatProto, hash, epoch)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMiniBlockByHash with proto")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"miniblock": miniBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONMiniBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrValidationEmptyBlockHash)
		return
	}

	epoch, err := getQueryParamEpoch(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetBlock, errors.ErrInvalidEpoch)
		return
	}

	start := time.Now()
	miniBlock, err := ib.getFacade().GetInternalMiniBlockByHash(common.ApiOutputFormatJSON, hash, epoch)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalMiniBlockByHash with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetBlock, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"miniblock": miniBlock}, "", shared.ReturnCodeSuccess)
}

func (ib *internalBlockGroup) getJSONStartOfEpochValidatorsInfo(c *gin.Context) {
	epoch, err := getQueryParamEpoch(c)
	if err != nil {
		shared.RespondWithValidationError(c, errors.ErrGetValidatorsInfo, errors.ErrInvalidEpoch)
		return
	}

	start := time.Now()
	validatorsInfo, err := ib.getFacade().GetInternalStartOfEpochValidatorsInfo(epoch)
	logging.LogAPIActionDurationIfNeeded(start, "API call: GetInternalStartOfEpochValidatorsInfo with JSON")
	if err != nil {
		shared.RespondWithInternalError(c, errors.ErrGetValidatorsInfo, err)
		return
	}

	shared.RespondWith(c, http.StatusOK, gin.H{"validators": validatorsInfo}, "", shared.ReturnCodeSuccess)
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

func getQueryParamEpoch(c *gin.Context) (uint32, error) {
	epochStr := c.Param("epoch")
	epoch, err := strconv.ParseUint(epochStr, 10, 32)
	return uint32(epoch), err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ib *internalBlockGroup) IsInterfaceNil() bool {
	return ib == nil
}
