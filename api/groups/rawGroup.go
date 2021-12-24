package groups

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	getRawBlockByNoncePath = "/block/by-nonce/:nonce"
	getRawBlockByHashPath  = "/block/by-hash/:hash"
	getRawBlockByRoundPath = "/block/by-round/:round"
)

// rawBlockFacadeHandler defines the methods to be implemented by a facade for handling block requests
type rawBlockFacadeHandler interface {
	GetRawBlockByHash(hash string, withTxs bool) (*block.MetaBlock, error)
	GetRawBlockByNonce(nonce uint64, withTxs bool) (*block.MetaBlock, error)
	GetRawBlockByRound(round uint64, withTxs bool) (*block.MetaBlock, error)
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
			Path:    getRawBlockByNoncePath,
			Method:  http.MethodGet,
			Handler: bg.getRawBlockByNonce,
		},
		{
			Path:    getRawBlockByHashPath,
			Method:  http.MethodGet,
			Handler: bg.getRawBlockByHash,
		},
		{
			Path:    getRawBlockByRoundPath,
			Method:  http.MethodGet,
			Handler: bg.getRawBlockByRound,
		},
	}
	bg.endpoints = endpoints

	return bg, nil
}

func (bg *rawBlockGroup) getRawBlockByNonce(c *gin.Context) {
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
	block, err := bg.getFacade().GetRawBlockByNonce(nonce, withTxs)
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

	//shared.RespondWith(c, http.StatusOK, gin.H{"block": block}, "", shared.ReturnCodeSuccess)

	c.ProtoBuf(
		http.StatusOK,
		block,
	)
}

func (bg *rawBlockGroup) getRawBlockByHash(c *gin.Context) {
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
	block, err := bg.getFacade().GetRawBlockByHash(hash, withTxs)
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

func (bg *rawBlockGroup) getRawBlockByRound(c *gin.Context) {
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
	block, err := bg.getFacade().GetRawBlockByRound(round, withTxs)
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
