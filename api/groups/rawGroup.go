package groups

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	getRawMetaBlockByNoncePath = "/metablock/by-nonce/:nonce"
	getRawMetaBlockByHashPath  = "/metablock/by-hash/:hash"
	getRawMetaBlockByRoundPath = "/metablock/by-round/:round"
)

// rawBlockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type rawBlockFacadeHandler interface {
	GetRawMetaBlockByHash(hash string, withTxs bool) ([]byte, error)
	GetRawMetaBlockByNonce(nonce uint64, withTxs bool) ([]byte, error)
	GetRawMetaBlockByRound(round uint64, withTxs bool) ([]byte, error)
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

	bg := &rawBlockGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    getRawMetaBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: bg.getRawMetaBlockByNonce,
		},
		{
			Path:    getRawMetaBlockByHashPath,
			Method:  http.MethodGet,
			Handler: bg.getRawMetaBlockByHash,
		},
		{
			Path:    getRawMetaBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: bg.getRawMetaBlockByRound,
		},
	}
	bg.endpoints = endpoints

	return bg, nil
}

func (bg *rawBlockGroup) getRawMetaBlockByNonce(c *gin.Context) {
	nonce, err := getQueryParamNonce(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetRawMetaBlockByNonce(nonce, withTxs)
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

	// c.ProtoBuf(
	// 	http.StatusOK,
	// 	block,
	// )
}

func (bg *rawBlockGroup) getRawMetaBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyBlockHash.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockNonce.Error()),
		)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetRawMetaBlockByHash(hash, withTxs)
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

	// c.ProtoBuf(
	// 	http.StatusOK,
	// 	shared.GenericAPIResponse{
	// 		Data:  gin.H{"block": block},
	// 		Error: "",
	// 		Code:  shared.ReturnCodeSuccess,
	// 	},
	// )
}

func (bg *rawBlockGroup) getRawMetaBlockByRound(c *gin.Context) {
	round, err := getQueryParamRound(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidBlockRound.Error()),
		)
		return
	}

	withTxs, err := getQueryParamWithTxs(c)
	if err != nil {
		shared.RespondWithValidationError(
			c, fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrInvalidQueryParameter.Error()),
		)
		return
	}

	start := time.Now()
	block, err := bg.getFacade().GetRawMetaBlockByRound(round, withTxs)
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

	shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

	// c.ProtoBuf(
	// 	http.StatusOK,
	// 	shared.GenericAPIResponse{
	// 		Data:  gin.H{"block": block},
	// 		Error: "",
	// 		Code:  shared.ReturnCodeSuccess,
	// 	},
	// )
}

func (ng *rawBlockGroup) getFacade() rawBlockFacadeHandler {
	ng.mutFacade.RLock()
	defer ng.mutFacade.RUnlock()

	return ng.facade
}

// UpdateFacade will update the facade
func (ng *rawBlockGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(rawBlockFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	ng.mutFacade.Lock()
	ng.facade = castFacade
	ng.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ng *rawBlockGroup) IsInterfaceNil() bool {
	return ng == nil
}
